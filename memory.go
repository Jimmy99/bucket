package distributed_token_bucket

import (
	"sync"
	"time"
	"errors"
)

type MemoryStorage struct {
	id int
	mutex sync.RWMutex
	buckets map[string]int
}

func (ms *MemoryStorage) Ping() error { return nil }

func (ms *MemoryStorage) Create(name string, capacity int) error {
	ms.id = 1

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

	if value, exists := ms.buckets[bucketName]; exists {
		if value >= tokens {
			ms.buckets[bucketName] -= tokens
			return nil
		}
	}

	return errors.New("Insufficient tokens.")
}

func (ms *MemoryStorage) Put(bucketName string, tokens int) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, exists := ms.buckets[bucketName]; exists {
		ms.buckets[bucketName] += tokens
		return nil
	}

	return errors.New("Bucket does not exist tokens.")
}

func (ms *MemoryStorage) Watch(bucketName string, tokens int, duration time.Duration) chan error {
	done := make(chan error)
	timeout := time.After(duration)

	go func(bucketName string, tokensDesired int, timeout <-chan time.Time, done chan error) {
		// time.Ticker returns a channel which fires every time the duration provided is passed
		ticker := time.NewTicker(time.Millisecond * 500)
		defer ticker.Stop()

		for {
			select {

			// attempt to take the desiredTokens on every ticker event
			case <-ticker.C:
				if err := ms.Take(bucketName, tokensDesired); err == nil {
					done <- nil
					break
				}

			// return an error if our timeout has passed
			case <-timeout:

				done <- errors.New("Watch timeout.")
				break
			}
		}
	}(bucketName, tokens, timeout, done)

	return done
}

