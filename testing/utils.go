package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mattesthaut/gdfs/core"
	"github.com/mattesthaut/gdfs/data"
)

const (
	basePort = 42000
)

func newNetwork(t *testing.T, size int) []*core.Host {
	t.Log("Création d'un réseau de", size, "noeuds")
	hosts := make([]*core.Host, size)

	for i := range size {
		addr := fmt.Sprintf("127.0.0.1:%d", basePort+i)
		storage := core.NewMemoryStorage()
		hosts[i] = core.NewHost(addr, storage)

		if err := hosts[i].Start(); err != nil {
			t.Fatalf("Erreur lors du démarrage du noeud %d: %v", i, err)
		}

	}

	time.Sleep(50 * time.Millisecond)

	t.Log("Phase de bootstrap des noeuds")
	for i, host := range hosts {
		connectedCount := 0

		if i > 0 {
			err := host.Bootstrap(hosts[i-1].Addr())
			if err == nil {
				connectedCount++
			}
		}

		if connectedCount == 0 && i != 0 {
			t.Fatalf("Le noeud %d n'a pas pu se connecter à aucun autre noeud", i)
		}
	}

	t.Log("Attente de stabilisation des tables de routage")
	time.Sleep(time.Second)

	return hosts
}

func destroyNetwork(hosts []*core.Host) {
	for _, host := range hosts {
		if host != nil {
			host.Stop()
		}
	}
}

func networkFailure(t *testing.T, hosts []*core.Host, ratio int) {
	disconnectedCount := 0
	for i := 0; i < len(hosts); i += ratio {
		hosts[i].Stop()
		hosts[i] = nil
		disconnectedCount++
	}
	t.Logf("Nombre de noeuds déconnectés: %d", disconnectedCount)

	t.Log("Attente de mise à jour des tables de routage")
	time.Sleep(3 * time.Second)
}

func networkStats(t *testing.T, hosts []*core.Host) {
	activeHosts := 0
	totalRtSize := 0
	minRtSize := 100000
	maxRtSize := 0

	for _, host := range hosts {
		if host != nil {
			rtSize := host.KnownPeerCount()
			totalRtSize += rtSize
			minRtSize = min(minRtSize, rtSize)
			maxRtSize = max(maxRtSize, rtSize)
			activeHosts++
		}
	}

	avgRtSize := float64(totalRtSize) / float64(activeHosts)

	t.Logf("Nombre de noeud actifs: %d", activeHosts)
	t.Logf("  - moyenne = %.1f", avgRtSize)
	t.Logf("  - minimum = %d", minRtSize)
	t.Logf("  - maximum = %d", maxRtSize)
}

func firstActiveHost(t *testing.T, hosts []*core.Host) int {
	for i := range hosts {
		if hosts[i] != nil {
			return i
		}
	}

	t.Fatal("Aucun noeud actif trouvé")
	return -1
}

func lastActiveHost(t *testing.T, hosts []*core.Host) int {
	for i := len(hosts) - 1; i >= 0; i-- {
		if hosts[i] != nil {
			return i
		}
	}

	t.Fatal("Aucun noeud actif trouvé")
	return -1
}

func store(t *testing.T, hosts []*core.Host, i int, d []byte) core.Id {
	t.Logf("Stockage de la donnée depuis le noeud %d", i)
	id, replicaCount := data.StoreData(d, hosts[i])

	if replicaCount == 0 {
		t.Fatal("Impossible de stocker la donnée sur le réseau")
	}

	t.Logf("Donnée aléatoire stockée :")
	t.Logf("  - id : %s", id.String())
	t.Logf("  - réplicas : %d", replicaCount)

	return id
}
