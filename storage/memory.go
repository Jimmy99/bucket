package storage

import (
	"sync"
	"errors"
)

// Implement an in-memory datastore with a concurrent safe map protected by a RWmutex
type MemoryStorage struct {
	mutex sync.RWMutex
	buckets map[string]int
}

func (ms *MemoryStorage) Ping() error { return nil }

// Create an entry in the map if one does not exist for the given name. If an entry already exists return nil so that
// the bucket will share the entry.
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

// Decrement the entry value unless the value < tokens. If value < tokens return an error else return nil.
// Note that we don't check if name exists as a key of the map, this is because storage.Take is called directly
// from it's parent method on Bucket where the name is derived from. A bucket can't exist to call if it doesn't already
// have a name...
func (ms *MemoryStorage) Take(bucketName string, tokens int) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if ms.buckets[bucketName] < tokens {
		return errors.New("Insufficient tokens.")
	}

	ms.buckets[bucketName] -= tokens
	return nil
}

// get and return the token value, set the token value to zero
func (ms *MemoryStorage) TakeAll(bucketName string) (int, error){
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	count := ms.buckets[bucketName]
	ms.buckets[bucketName] = 0

	return count, nil
}

func (ms *MemoryStorage) Set(bucketName string, tokens int) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.buckets[bucketName] = tokens
	return nil
}

// Increment the entry value by the given tokens integer
func (ms *MemoryStorage) Put(bucketName string, tokens int) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.buckets[bucketName] += tokens
	return nil
}

// We only need a read lock here because we are only reading it.
func (ms *MemoryStorage) Count(bucketName string) (int, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	return ms.buckets[bucketName], nil
}