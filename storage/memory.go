package storage

import (
	"sync"
	"errors"
)

type MemoryStorage struct {
	mutex sync.RWMutex
	buckets map[string]int
}

func (ms *MemoryStorage) Ping() error { return nil }

func (ms *MemoryStorage) Create(name string, capacity int) error {
	if ms.buckets == nil {
		ms.buckets = map[string]int{}
	}

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, exists := ms.buckets[name]; exists {
		return nil
	}

	ms.buckets[name] = capacity

	return nil
}

func (ms *MemoryStorage) Take(bucketName string, tokens int) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if ms.buckets[bucketName] < tokens {
		return errors.New("Insufficient tokens.")
	}

	ms.buckets[bucketName] -= tokens
	return nil
}

func (ms *MemoryStorage) Put(bucketName string, tokens int) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.buckets[bucketName] += tokens
	return nil
}

func (ms *MemoryStorage) Count(bucketName string) (int, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	return ms.buckets[bucketName], nil
}