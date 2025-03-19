package core

import (
	"time"
)

const (
	IdSize    = 20   // taille des identifiants en octet
	ValueSize = 1024 // taille d'une valeur en octet

	// nombre de noeuds à interroger à chaque itération de FindNode et FindValue
	batchSize        = 3
	maxReplicasCount = 5

	bucketCapacity = 20 // nombre maximum de noeuds connus = 8*IdSize*bucketCapacity

	storageTtl      = 60 * time.Minute // durée de vie d'une valeur
	storageCapacity = 64 * 1024        // nombre de valeurs maximal

	connTtl = 3 * time.Second // durée de vie maximale d'une connexion

	cleanupFreq = time.Minute * 10 // fréquence de nettoyage de la table de routage
)
