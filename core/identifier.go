package core

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
)

// Id est un identifiant pour un noeud ou une valeur du réseau.
type Id [IdSize]byte

// Retourne un identifiant aléatoire.
func NewRandomId() Id {
	var id Id
	_, _ = rand.Read(id[:])
	return id
}

// Hash une donnée et retourne son identifiant associé.
func NewIdFrom(data []byte) Id {
	var id Id
	hash := sha1.Sum(data)
	copy(id[:], hash[:])
	return id
}

// Crée un identifiant à partir de sa représentation hexadécimale.
func IdFromString(str string) (Id, error) {
	var id Id
	d, err := hex.DecodeString(str)
	if err != nil {
		return id, err
	}

	copy(id[:], d)
	return id, nil
}

func (id Id) Equal(other Id) bool {
	return bytes.Equal(id[:], other[:])
}

func (id Id) Less(other Id) bool {
	return bytes.Compare(id[:], other[:]) < 0
}

// Retourne la représentation hexadécimale d'un identifiant.
func (id Id) String() string {
	return hex.EncodeToString(id[:])
}

// Calcule la distance XOR entre deux identifiants.
func (id Id) Distance(other Id) Id {
	var dist Id

	for i := range IdSize {
		dist[i] = id[i] ^ other[i]
	}

	return dist
}

// Retourne le nombre de bits successifs en commun à partir du poids fort.
func (id Id) PrefixLen(other Id) int {
	count := 0

	for i := range IdSize {
		if id[i] == other[i] {
			count += 8
			continue
		}

		xor := id[i] ^ other[i]
		for j := range 8 {
			if (xor & (1 << (7 - j))) != 0 {
				return count
			}
			count++
		}
	}

	return count
}
