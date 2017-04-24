package distributed_token_bucket

import (
	"time"
)

/**
 * bucket.go provides the operations for a basic distributed and persistent token-bucket.
 *
 * It allows multiple processes to share the same bucket via sharing an identical key and database instance.
 * Notably this implementation of a token bucket does not refill itself on a given interval but instead relies on
 * clients to refill tokens when they are done being used. The reason for this is to eliminate contention and
 * duplication i.e. multiple clients trying to refill the bucket on their own system-dependant time interval.
 *
 * It should be realized that if a client does not responsibly refill its tokens after use there will be no tokens left
 * for other clients.
 */

type Bucket struct {
	storage IStorage

	// the name of the bucket, used as the key in Redis for which the token value is stored
	Name string

	// the token value a bucket should hold when it is created, if the bucket already exists this does nothing
	capacity int
}

// Connect and verify a redis client connection while instantiating a bucket then create a key-value pair in the database if one does not already exist for the given name
func NewBucket(name string, capacity int, storage IStorage) (*Bucket, error) {
	bucket := &Bucket{Name: name, capacity: capacity, storage: storage}

	// ensure our redis connection is valid
	err := bucket.storage.Ping()
	if err != nil {
		return nil, err
	}

	err = bucket.Create(name, capacity)

	return bucket, err
}

// bucket.Create will create a new bucket with the given parameters if one does not exist, if no bucket can be created it will return an error
func (bucket *Bucket) Create(name string, capacity int) error {
	return bucket.storage.Create(name, capacity)
}

// Executes a lua script which decrements the token value by tokensDesired if tokensDesired >= the token value.
func (bucket *Bucket) Take(tokensDesired int) error {
	return bucket.storage.Take(bucket.Name, tokensDesired)
}

// Increment the token value by a given amount
func (bucket *Bucket) Put(amount int) error {
	return bucket.storage.Put(bucket.Name, amount)
}

// attempt on a 500ms interval to asynchronously call bucket.Take until timeout is exceeded
// returns a channel which will fire nil or error on completion
func (bucket *Bucket) Watch(tokensDesired int, duration time.Duration) chan error {
	return bucket.storage.Watch(bucket.Name, tokensDesired, duration)
}
