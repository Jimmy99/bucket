package distributed_token_bucket

import (
	"github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
	"time"
	"errors"

	"fmt"
)

// Use redis-lock to transactionally access and act on information which could be in contention from concurrent access.
//
// Used for logic such as:
//
// GET my_bucket
// if my_bucket > some_amount then DECR my_bucket some_amount
// else return error
//
// There might be a lock-less way to do it but I couldn't think of a simple implementation for it.
var REDIS_LOCK_OPTIONS = &lock.LockOptions{ WaitTimeout: time.Second, WaitRetry: time.Millisecond * 50}

// create a superset struct of redis.Client to provide our token-bucket primitives
type Storage struct {
	*redis.Client
}

func NewStorage(options *redis.Options) *Storage {
	client := redis.NewClient(options)
	return &Storage{client}
}

func (storage *Storage) Create(name string, capacity int) error {
	// query redis and see if the key already exists
	_, err := storage.Get(name).Result()

	// if redis returns an error AND that error is not for a key that does not exist return that error
	if err != nil && err != redis.Nil {
		return err
	}

	// if the error is for a key that does not exist create a key with a value of capacity
	if err == redis.Nil {
		// the last param 0 indicates the key will never expire
		if _, err = storage.Set(name, capacity, 0).Result(); err != nil {
			return err
		}
	}

	// great we know the bucket exists and has been created if it doesn't already exist
	return nil
}


// Because this method may be called concurrently and the token value of the requested bucket could be decremented
// in the middle of an operation, use a lock to transactionally check if enough tokens exist then if enough tokens exist
// decrement the tokens value by tokensDesired.
//
// This is not an ideal approach if you have a lot of instances trying to take from the same bucket at once.
func (storage *Storage) Take(name string, tokensDesired int) error {
	lockKey := fmt.Sprintf("%s.lock", name)

	// ObtainLock will block if a lock cannot be immediately acquired with WaitTimeout and WaitRetry defined by REDIS_LOCK_OPTIONS
	dbLock, err := lock.ObtainLock(storage, lockKey, REDIS_LOCK_OPTIONS)
	defer dbLock.Unlock() // don't forget to return the lock

	if err != nil {
		return err
	}

	tokensCount, err := storage.Get(name).Int64()

	if err != nil {
		return err
	}

	// check if enough tokens exist to satisfy tokensDesired
	if tokensCount < int64(tokensDesired) {
		return errors.New("Insufficient tokens.")
	}

	// finally decrement by tokensDesired
	return storage.DecrBy(name, int64(tokensDesired)).Err()
}

// Increment the tokens value by the given amount. A lock is not necessary because a 'Insufficient tokens.', which may
// occur if we don't use a lock, is entirely acceptable.
//
// For example two concurrent processes may cause the following order of operations to occur:
// GET my_bucket => 10
// INCR my_bucket 10 => 20
// DECR my_bucket 15 => errors.New("Insufficient tokens.")
//
// and that is intended to be OK
func (storage *Storage) Put(name string, amount int) error {
	return storage.IncrBy(name, int64(amount)).Err()
}