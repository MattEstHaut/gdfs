package data

import (
	"sync"

	"github.com/mattesthaut/gdfs/core"
)

const (
	// Nombre maximal de requêtes parallèles par ParallelReader
	// et ParallelWriter
	semInitialValue = 20
)

type Reader interface {
	// FindValue doit être est sûre pour une utilisation concurrente.
	FindValue(id core.Id) (core.Value, bool)
}

type Writer interface {
	// StoreValue doit être est sûre pour une utilisation concurrente.
	StoreValue(value core.Value) (core.Id, int)
}

// Un ParallelReader permet de paralléliser les opérations
// de lecture sur un Reader.
type ParallelReader struct {
	reader Reader
	sem    chan struct{}
}

// Un ParallelWriter permet de paralléliser les opérations
// d'écriture sur un Writer.
type ParallelWriter struct {
	writer Writer
	sem    chan struct{}
}

// Crée un ParallelReader à partir d’un Reader sûr pour une
// utilisation concurrente.
func NewParallelReader(reader Reader) *ParallelReader {
	return &ParallelReader{
		reader: reader,
		sem:    make(chan struct{}, semInitialValue),
	}
}

// Crée un ParallelWriter à partir d’un Writer sûr pour une
// utilisation concurrente.
func NewParallelWriter(writer Writer) *ParallelWriter {
	return &ParallelWriter{
		writer: writer,
		sem:    make(chan struct{}, semInitialValue),
	}
}

// Cherche les valeurs associées aux identifiants et les retourne
// dans le même ordre. La deuxième valeur de retour est true si
// et seulement si toutes les valeurs ont été trouvées.
func (pr *ParallelReader) FindValues(ids []core.Id) ([]core.Value, bool) {
	results := make([]core.Value, len(ids))

	var wg sync.WaitGroup
	allFound := true

	for i, id := range ids {
		wg.Add(1)

		go func(index int, id core.Id) {
			defer wg.Done()

			pr.sem <- struct{}{}
			defer func() { <-pr.sem }()

			if value, found := pr.reader.FindValue(id); found {
				results[i] = value
			} else {
				allFound = false
			}
		}(i, id)
	}

	wg.Wait()
	return results, allFound
}

// Stocke les paires identifiant-valeur et retourne true si
// toutes les paires ont été stockées, sinon false.
func (pr *ParallelWriter) StoreValues(values []core.Value) int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	replicas := 1000

	for _, value := range values {
		wg.Add(1)

		go func(value core.Value) {
			defer wg.Done()

			pr.sem <- struct{}{}
			defer func() { <-pr.sem }()

			_, r := pr.writer.StoreValue(value)

			mu.Lock()
			defer mu.Unlock()
			replicas = min(replicas, r)
		}(value)
	}

	wg.Wait()
	return replicas
}
