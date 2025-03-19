package core

import (
	"sort"
)

// Peer est un noeud distant.
type Peer struct {
	Id   Id
	Addr string // adresse physique du noeud
}

// Trie les noeuds par distance croissante d’un identifiant.
func sortPeersByDistance(peers []Peer, id Id) {
	sort.Slice(peers, func(i int, j int) bool {
		return peers[i].Id.Distance(id).Less(peers[j].Id.Distance(id))
	})
}

// Retourne les n premiers noeuds de la liste. Si la liste est trop
// courte, elle est retournée entièrement.
func firstNPeers(peers []Peer, n int) []Peer {
	if len(peers) < n {
		return peers
	}

	return peers[:n]
}
