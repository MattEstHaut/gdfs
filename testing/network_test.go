package test

import (
	"crypto/rand"
	"slices"
	"testing"

	"github.com/mattesthaut/gdfs/core"
	"github.com/mattesthaut/gdfs/data"
)

const (
	nodeCount       = 200
	disconnectRatio = 4
	randomDataSize  = core.ValueSize * 100
	testFileSrc     = "test.txt"
	testFileDest    = "test_out.txt"
)

func TestRandomValue(t *testing.T) {
	hosts := newNetwork(t, nodeCount)
	defer destroyNetwork(hosts)

	networkStats(t, hosts)

	t.Logf("Création d'une donnée aléatoire de %d octets", randomDataSize)
	randomData := make([]byte, randomDataSize)
	if _, err := rand.Read(randomData[:]); err != nil {
		t.Fatalf("Erreur lors de la création de la donnée de test: %v", err)
	}

	sourceNodeId := firstActiveHost(t, hosts)
	id := store(t, hosts, sourceNodeId, randomData)

	networkFailure(t, hosts, disconnectRatio)
	networkStats(t, hosts)

	retrievalNodeId := lastActiveHost(t, hosts)

	t.Logf("Récupération de la donnée aléatoire depuis le noeud %d", retrievalNodeId)
	retrievedData, found := data.FindData(id, hosts[retrievalNodeId])

	if !found {
		t.Error("La donnée n'a pas été trouvée")
	} else {
		if !slices.Equal(retrievedData, randomData) {
			t.Error("La donnée récupérée ne correspond pas à l'original")
		} else {
			t.Log("La donnée a été récupérée et correspond à l'original")
		}
	}
}

func TestWithFile(t *testing.T) {
	hosts := newNetwork(t, nodeCount)
	defer destroyNetwork(hosts)

	networkStats(t, hosts)

	t.Logf("Récupération du fichier %s", testFileSrc)
	fileData, err := data.ReadFile(testFileSrc)
	if err != nil {
		t.Fatalf("Impossible de lire %s", testFileDest)
	}

	sourceNodeId := firstActiveHost(t, hosts)
	id := store(t, hosts, sourceNodeId, fileData)

	networkFailure(t, hosts, disconnectRatio)
	networkStats(t, hosts)

	retrievalNodeId := lastActiveHost(t, hosts)

	t.Logf("Récupération du fichier depuis le noeud %d", retrievalNodeId)
	retrievedData, found := data.FindData(id, hosts[retrievalNodeId])

	if !found {
		t.Error("Le fichier n'a pas été trouvée")
	} else {
		if !slices.Equal(retrievedData, fileData) {
			t.Error("Le fichier récupéré ne correspond pas à l'original")
		} else {
			t.Log("Le fichier a été récupéré et correspond à l'original")
		}
	}

	if err := data.WriteFile(testFileDest, retrievedData); err != nil {
		t.Errorf("Impossible d'écrire les données dans %s", testFileDest)
	} else {
		t.Logf("Les données ont été écrites dans %s", testFileDest)
	}
}
