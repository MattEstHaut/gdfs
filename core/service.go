// Le package core implémente une table de hachage distribuée et
// les utilitaires pour créer des réseaux, retrouver des données
// et stocker des données. Les noeuds maintiennent
// automatiquement une table de routage.
package core

// Retrouve les bucketCapacity noeuds les plus proches de target.
func (h *Host) FindNode(target Id) []Peer {
	closestPeers := h.closestPeersFrom(target, bucketCapacity)
	visitedPeers := make(peerSet)
	discoveredPeers := make(peerSet)

	discoveredPeers.addMany(closestPeers)

	for {
		batch := nextBatch(closestPeers, visitedPeers)
		if len(batch) == 0 {
			break
		}
		visitedPeers.addMany(batch)

		for _, peer := range batch {
			if newPeers, err := h.findNodeFrom(peer.Addr, target); err == nil {
				for _, newPeer := range newPeers {
					if !discoveredPeers.has(newPeer) {
						closestPeers = append(closestPeers, newPeer)
						discoveredPeers.add(newPeer)

						if !newPeer.Id.Equal(h.id) {
							h.rt.addPeer(newPeer)
						}
					}
				}
			}
		}

		sortPeersByDistance(closestPeers, target)
		closestPeers = firstNPeers(closestPeers, bucketCapacity)
	}

	return closestPeers
}

// Retrouve la valeur associée à l'identifiant. La deuxième valeur
// de retour est true si et seulement si la donnée a été retrouvée.
func (h *Host) FindValue(id Id) (Value, bool) {
	closestPeers := h.closestPeersFrom(id, bucketCapacity)
	visitedPeers := make(peerSet)
	discoveredPeers := make(peerSet)

	discoveredPeers.addMany(closestPeers)

	for {
		batch := nextBatch(closestPeers, visitedPeers)
		if len(batch) == 0 {
			break
		}
		visitedPeers.addMany(batch)

		for _, peer := range batch {
			if res, err := h.findValueFrom(peer.Addr, id); err == nil {
				if res.Found {
					return res.Value, true
				}

				for _, newPeer := range res.Nodes {
					if !discoveredPeers.has(newPeer) {
						closestPeers = append(closestPeers, newPeer)
						discoveredPeers.add(newPeer)

						if !newPeer.Id.Equal(h.id) {
							h.rt.addPeer(newPeer)
						}
					}
				}
			}
		}

		sortPeersByDistance(closestPeers, id)
		closestPeers = firstNPeers(closestPeers, bucketCapacity)
	}

	return Value{}, false
}

// Stocke la valeur et renvoie son identifiant. La deuxième valeur
// de retour est le nombre de replicas qui ont été stockés. Si le
// nombre de replicas est nul, alors la donnée n’a pas été
// correctement stockée.
func (h *Host) StoreValue(value Value) (Id, int) {
	id := NewIdFrom(value[:])

	peers := h.FindNode(id)
	replicasCount := 0

	for _, peer := range peers {
		if ok, err := h.storeTo(peer.Addr, id, value); err == nil && ok {
			replicasCount++
			if replicasCount >= maxReplicasCount {
				break
			}
		}
	}

	return id, replicasCount
}

type peerSet map[Peer]struct{}

func (s peerSet) addMany(peers []Peer) {
	for _, peer := range peers {
		s.add(peer)
	}
}

func (s peerSet) add(peer Peer) {
	s[peer] = struct{}{}
}

func (s peerSet) has(peer Peer) bool {
	_, ok := s[peer]
	return ok
}

func nextBatch(peers []Peer, visited peerSet) []Peer {
	batch := make([]Peer, 0, batchSize)

	for _, peer := range peers {
		if !visited.has(peer) {
			batch = append(batch, peer)
			if len(batch) >= batchSize {
				return batch
			}
		}
	}

	return batch
}
