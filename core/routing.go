package core

import (
	"slices"
	"sync"
)

type bucket []Peer

// routingTable est la table de routage du noeud local. Elle contient
// ses noeuds connus. Elle fonctionne de manière à connaître mieux
// son voisinage que les noeuds plus éloignés. routingTable est sûre
// pour une utilisation concurrente.
type routingTable struct {
	buckets [IdSize * 8]bucket
	id      Id // l'identifiant du noeud local
	mu      sync.Mutex
	cancel  chan struct{}
}

func newRoutingTable(id Id) *routingTable {
	rt := routingTable{
		id:     id,
		cancel: make(chan struct{}),
	}

	for i := range rt.buckets {
		rt.buckets[i] = make(bucket, 0, bucketCapacity)
	}

	return &rt
}

func (rt *routingTable) addPeer(peer Peer) bool {
	bucket := rt.getBucketOf(peer.Id)

	rt.mu.Lock()
	defer rt.mu.Unlock()

	if len(*bucket) >= bucketCapacity {
		return false
	}

	for _, existing := range *bucket {
		if existing.Id.Equal(peer.Id) {
			return false
		}
	}

	*bucket = append(*bucket, peer)
	return true
}

func (rt *routingTable) removePeer(id Id) bool {
	bucket := rt.getBucketOf(id)

	rt.mu.Lock()
	defer rt.mu.Unlock()

	for i, peer := range *bucket {
		if peer.Id.Equal(id) {
			*bucket = slices.Delete(*bucket, i, i+1)
			return true
		}
	}

	return false
}

func (rt *routingTable) peers() []Peer {
	peers := make([]Peer, 0)

	rt.mu.Lock()
	defer rt.mu.Unlock()

	for _, bucket := range rt.buckets {
		peers = append(peers, bucket[:]...)
	}

	return peers
}

func (rt *routingTable) getBucketOf(id Id) *bucket {
	prefixLen := rt.id.PrefixLen(id)
	index := IdSize*8 - prefixLen - 1
	return &rt.buckets[index]
}
