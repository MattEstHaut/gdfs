package core

import (
	"encoding/gob"
	"net"
	"time"
)

const (
	// Description des types de requête plus bas.
	PingRequestType = iota
	FindNodeRequestType
	FindValueRequestType
	StoreRequestType
)

// Request est une requête envoyer d'un noeud à un autre. Elle
// a une taille fixe pour tous les types de requête.
type Request struct {
	Type       int
	Id         Id
	Value      Value
	SenderAddr string
	SenderId   Id
}

type findValueResponse struct {
	Found bool
	Value Value
	Nodes []Peer
}

func newPingRequest() Request {
	return Request{
		Type: PingRequestType,
	}
}

func newFindNodeRequest(target Id) Request {
	return Request{
		Type: FindNodeRequestType,
		Id:   target,
	}
}

func newFindValueRequest(target Id) Request {
	return Request{
		Type: FindValueRequestType,
		Id:   target,
	}
}

func newStoreRequest(id Id, value Value) Request {
	return Request{
		Type:  StoreRequestType,
		Id:    id,
		Value: value,
	}
}

func (r Request) sign(addr string, id Id) Request {
	r.SenderAddr = addr
	r.SenderId = id
	return r
}

// Demande l'identifiant d'un noeud.
func (h *Host) pingPeer(addr string) (Id, error) {
	return requestTo[Id](addr, newPingRequest().sign(h.addr, h.id))
}

// Demande les bucketCapacity noeuds les plus proches de target de la table
// de routage du noeud[addr].
func (h *Host) findNodeFrom(addr string, target Id) ([]Peer, error) {
	req := newFindNodeRequest(target).sign(h.addr, h.id)
	return requestTo[[]Peer](addr, req)
}

// Demande la valeur de clé target. Si le noeud[addr] ne l'a pas, il répond avec
// les bucketCapacity noeuds les plus proches de target de sa table de routage.
func (h *Host) findValueFrom(addr string, target Id) (findValueResponse, error) {
	req := newFindValueRequest(target).sign(h.addr, h.id)
	return requestTo[findValueResponse](addr, req)
}

// Demande à stocker la pair key-value sur le noeud[addr].
func (h *Host) storeTo(addr string, key Id, value Value) (bool, error) {
	req := newStoreRequest(key, value).sign(h.addr, h.id)
	return requestTo[bool](addr, req)
}

func requestTo[R any](addr string, req Request) (R, error) {
	var ret R

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return ret, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(connTtl))

	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	if encoder.Encode(req) != nil {
		return ret, err
	}

	err = decoder.Decode(&ret)

	return ret, err
}
