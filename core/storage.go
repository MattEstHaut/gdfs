package core

import (
	"sync"
	"time"
)

// Value est une donnée stockée sur le réseau.
type Value [ValueSize]byte

// Un Storage permet de stocker des pairs identifiant-valeur sur le noeud local.
type Storage interface {
	Get(id Id) (Value, bool)
	Set(id Id, value Value) bool
}

// MemoryStorage est un Storage en mémoire.
type MemoryStorage struct {
	data   map[Id]ValueWithExpiry
	mu     sync.Mutex
	size   int
	cancel chan struct{}
	wg     sync.WaitGroup
}

type ValueWithExpiry struct {
	Value    Value
	ExpireAt time.Time
}

func NewMemoryStorage() *MemoryStorage {
	ms := &MemoryStorage{
		data:   make(map[Id]ValueWithExpiry),
		cancel: make(chan struct{}),
	}

	ms.wg.Add(1)
	go ms.cleanupExpired()

	return ms
}

// Arrête le processus de suppression des valeurs expirées.
func (s *MemoryStorage) Close() {
	close(s.cancel)
	s.wg.Wait()
}

func (s *MemoryStorage) Get(id Id) (Value, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[id]
	if !exists {
		return Value{}, false
	}

	return item.Value, true
}

func (s *MemoryStorage) Set(id Id, value Value) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.size >= storageCapacity {
		return false
	}

	s.data[id] = ValueWithExpiry{
		Value:    value,
		ExpireAt: time.Now().Add(storageTtl),
	}

	s.size += 1

	return true
}

func (s *MemoryStorage) Len() int {
	return s.size
}

// FakeStorage est un Storage qui refuse de stocker des données.
type FakeStorage struct{}

func NewFakeStorage() *FakeStorage {
	return &FakeStorage{}
}

func (*FakeStorage) Get(id Id) (Value, bool) {
	return Value{}, false
}

func (*FakeStorage) Set(id Id, value Value) bool {
	return false
}
func (s *MemoryStorage) cleanupExpired() {
	defer s.wg.Done()

	ticker := time.NewTicker(storageTtl / 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()

			for id, item := range s.data {
				if now.After(item.ExpireAt) {
					delete(s.data, id)
					s.size -= 1
				}
			}

			s.mu.Unlock()

		case <-s.cancel:
			return
		}
	}
}
