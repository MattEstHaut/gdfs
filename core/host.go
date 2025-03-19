package core

import (
	"context"
	"encoding/gob"
	"net"
	"sync"
	"time"
)

// Host est le noeud local.
type Host struct {
	id      Id
	addr    string // l'adresse physique d'écoute
	storage Storage

	requests chan Request
	listener net.Listener

	rt routingTable

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewHost(addr string, storage Storage) *Host {
	id := NewRandomId()
	ctx, cancel := context.WithCancel(context.Background())

	return &Host{
		id:       id,
		addr:     addr,
		storage:  storage,
		requests: make(chan Request, 1024),
		rt:       *newRoutingTable(id),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (h *Host) Addr() string {
	return h.addr
}

func (h *Host) Id() Id {
	return h.id
}

func (h *Host) KnownPeerCount() int {
	return len(h.rt.peers())
}

// Connecte le noeud local à un réseau à partir d'un noeud
// distant y appartenant.
func (h *Host) Bootstrap(addr string) error {
	id, err := h.pingPeer(addr)
	if err != nil {
		return err
	}

	h.rt.addPeer(Peer{
		Id:   id,
		Addr: addr,
	})

	h.FindNode(h.id)

	return nil
}

// Écoute et répond aux autres noeuds du réseau.
func (h *Host) Start() error {
	h.startCleanup()
	return h.listen()
}

func (h *Host) Stop() {
	if h.listener != nil {
		h.cancel()
		h.listener.Close()
		h.listener = nil
		h.wg.Wait()
	}
}

func (h *Host) listen() error {
	listener, err := net.Listen("tcp", h.addr)
	if err != nil {
		return err
	}

	h.listener = listener

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		for {
			select {
			case <-h.ctx.Done():
				return
			default:
				if conn, err := listener.Accept(); err == nil {
					h.wg.Add(1)
					go h.handleConn(conn)
				}
			}
		}
	}()

	return nil
}

func (h *Host) handleConn(conn net.Conn) {
	defer h.wg.Done()
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(connTtl))

	var req Request

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	if decoder.Decode(&req) == nil {
		encoder.Encode(h.handleReq(req))
	}
}

func (h *Host) handleReq(req Request) any {
	if !req.SenderId.Equal(h.id) && req.SenderAddr != "" {
		h.rt.addPeer(Peer{
			Id:   req.SenderId,
			Addr: req.SenderAddr,
		})
	}

	switch req.Type {
	case PingRequestType:
		return h.id

	case FindNodeRequestType:
		return h.closestPeersFrom(req.Id, bucketCapacity)

	case FindValueRequestType:
		res := findValueResponse{}
		if val, ok := h.storage.Get(req.Id); ok {
			res.Found = true
			res.Value = val
		} else {
			res.Found = false
			res.Nodes = h.closestPeersFrom(req.Id, bucketCapacity)
		}
		return res

	case StoreRequestType:
		return h.storage.Set(req.Id, req.Value)

	default:
		return struct{}{}
	}
}

func (h *Host) closestPeersFrom(id Id, n int) []Peer {
	peers := h.rt.peers()
	sortPeersByDistance(peers, id)
	return firstNPeers(peers, n)
}

func (h *Host) startCleanup() {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		ticker := time.NewTicker(cleanupFreq)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.cleanup()
			case <-h.ctx.Done():
				return
			}
		}
	}()
}

// Retire les noeuds ne répondant pas de la table de routage
func (h *Host) cleanup() {
	peers := h.rt.peers()

	for _, peer := range peers {
		h.wg.Add(1)

		go func() {
			defer h.wg.Done()

			id, err := h.pingPeer(peer.Addr)
			if err != nil {
				h.rt.removePeer(peer.Id)
				return
			}

			if id != peer.Id {
				h.rt.removePeer(peer.Id)
				h.rt.addPeer(Peer{
					Id:   id,
					Addr: peer.Addr,
				})
			}
		}()
	}
}
