// Le package data implémente des utilitaires pour stocker et retrouver
// des données de taille quelconque sur un réseau dfsgo.
// Les données sont représentées sous forme d’arbre sur le réseau, où
// chaque feuille contient une partition de la donnée et chaque noeud
// interne contient une liste ordonnée des identifiants de ses enfants.
package data

import (
	"encoding/binary"

	"github.com/mattesthaut/gdfs/core"
)

const (
	headerSize  = 5                         // Taille de l’entête d’un noeud
	payloadSize = core.ValueSize - 5        // Taille maximale d’une donnée dans un noeud
	maxChildren = payloadSize / core.IdSize // Nombre maximum d’enfants d’un noeud interne
)

// chunk est une partition de donnée.
type chunk [payloadSize]byte

// Un node représente soit une partition de donnée, soit une liste
// ordonnée d’identifiants vers ses noeuds enfants. Son entête est
// composée de isLeaf et size.
type node struct {
	isLeaf   bool  // 1 octet
	size     int32 // 4 octets
	chunk    chunk
	children []*node
}

// Retrouve et renvoie une donnée de taille quelconque à partir
// de son identifiant. La deuxième valeur de retour est true si
// et seulement si la donnée a été intégralement retrouvée.
func FindData(id core.Id, reader Reader) ([]byte, bool) {
	pr := NewParallelReader(reader)
	if root, found := reader.FindValue(id); found {
		return join(root, pr)
	}
	return []byte{}, false
}

// Stocke une donnée de taille quelconque et renvoie son identifiant.
// La deuxième valeur de retour est le nombre de replicas qui ont été
// stockés. Si le nombre de replicas est nul, alors la donnée n’a pas
// été correctement stockée.
func StoreData(data []byte, writer Writer) (core.Id, int) {
	pw := NewParallelWriter(writer)
	id, values := Split(data)
	return id, pw.StoreValues(values)
}

// Découpe une donnée de taille quelconque en un arbre et renvoie
// l’identifiant de la racine et la liste des noeuds sous forme de Value.
func Split(data []byte) (core.Id, []core.Value) {
	leafs := splitIntoLeaf(data)
	tree := buildTree(leafs)
	return createValuesFromTree(tree)
}

// Reconstruit la donnée initiale à partir d’un arbre sous forme de
// liste de Value. La deuxième valeur de retour représente la
// réussite de l’opération.
func join(value core.Value, reader *ParallelReader) ([]byte, bool) {
	isLeaf := value[0] == 1
	size := int32(binary.BigEndian.Uint32(value[1:5]))

	if isLeaf {
		return value[headerSize : headerSize+size], true
	}

	ids := make([]core.Id, size)
	for i := range int(size) {
		s := headerSize + i*core.IdSize
		copy(ids[i][:], value[s:s+core.IdSize])
	}

	values, found := reader.FindValues(ids)
	if !found {
		return []byte{}, false
	}

	data := make([]byte, 0)
	for i := range values {
		childData, found := join(values[i], reader)
		if !found {
			return []byte{}, false
		}

		data = append(data, childData...)
	}

	return data, true
}

// Construit les noeuds internes de l’arbre à partir des feuilles.
func buildTree(leafs []node) *node {
	if len(leafs) == 1 {
		return &leafs[0]
	}

	parents := make([]node, 0)

	for i := 0; i < len(leafs); i += maxChildren {
		j := min(i+maxChildren, len(leafs))
		group := leafs[i:j]

		parent := node{
			isLeaf:   false,
			size:     int32(j - i),
			children: make([]*node, 0),
		}

		for i := range group {
			parent.children = append(parent.children, &group[i])
		}

		parents = append(parents, parent)
	}

	return buildTree(parents)
}

// Convertit un arbre de node en liste de Value. La première
// valeur de retour est l’identifiant de la racine.
func createValuesFromTree(root *node) (core.Id, []core.Value) {
	values := make([]core.Value, 0)

	value := core.Value{}
	binary.BigEndian.PutUint32(value[1:5], uint32(root.size))

	if root.isLeaf {
		value[0] = 1
		copy(value[5:core.ValueSize], root.chunk[:])
	} else {
		value[0] = 0

		for i, child := range root.children {
			childId, childValues := createValuesFromTree(child)
			s := 5 + i*core.IdSize
			copy(value[s:s+core.IdSize], childId[:])
			values = append(values, childValues...)
		}
	}

	values = append(values, value)
	return core.NewIdFrom(value[:]), values
}

// Découpe une donnée en feuilles.
func splitIntoLeaf(data []byte) []node {
	leafs := make([]node, 0)

	for i := 0; i < len(data); i += payloadSize {
		var ch chunk
		j := min(i+payloadSize, len(data))
		copy(ch[:], data[i:j])

		leafs = append(leafs, node{
			isLeaf: true,
			size:   int32(j - i),
			chunk:  ch,
		})
	}

	return leafs
}
