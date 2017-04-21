package distributed_token_bucket

import (
	"github.com/go-redis/redis"
)

/**
 * bucket.go provides the operations for a basic distributed and persistent token-bucket.
 *
 * It is intended to be used in a distributed system where multiple instances can share the same
 * bucket via an identical name parameter and database instance. Notably this implementation of a token bucket
 * does not refill itself on a given interval but instead relies on clients to refill tokens when they are done
 * being used. The reason for this is to eliminate contention and duplication i.e. multiple clients trying to refill
 * the bucket on their own system-dependant interval.
 *
 * It should be realized that if a client does not responsibly refill its tokens after use there will be no tokens left
 * for other clients.
 */

type Bucket struct {
	// the name of the bucket, used as the key in Redis from which the token value is stored
	name string

	// the token value a bucket should hold when it is created, if the bucket already exists this is not used
	capacity int

	// the storage backend, an extended version of the redis.Client object
	storage *Storage
}

func NewBucket(name string, capacity int, storageOptions *redis.Options) (*Bucket, error) {
	storage := NewStorage(storageOptions)

	bucket := &Bucket{ name: name, capacity: capacity, storage: storage }

	// ensure our redis connection is valid
	_, err := bucket.Ping()
	if err != nil {

		return nil, err
	}

	// bucket.storage.Create will create a new bucket with the given parameters if one does not exist
	err = bucket.storage.Create(name, capacity)

	return bucket, err
}

// simple ping-pong test to verify redis connection is solid
func (bucket *Bucket) Ping() (string, error){
	return bucket.storage.Ping().Result()
}

// Attempt to take some amount of tokens from the bucket and if that quantity cannot be taken return an error with the
// expectation of bucket.Take() being recalled.
//
// The return of errors.New("Insufficient tokens.") should be handled gracefully.
//
// Safe for concurrent access via a redis-based lock system.
func (bucket *Bucket) Take(amount int) error {
	return bucket.storage.Take(bucket.name, amount)
}

// Increment the token value of bucket by the given amount. If accessed concurrently may cause bucket.Take to return
// errors.New("Insufficient tokens.") inaccurately.
//
// This library intends "Insufficient tokens." to be handled gracefully so I still consider bucket.Put to be
// safe for concurrent access.
func (bucket *Bucket) Put(amount int) error {
	return bucket.storage.Put(bucket.name, amount)
}