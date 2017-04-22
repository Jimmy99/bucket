package distributed_token_bucket

import (
	"time"
	"errors"
	"github.com/go-redis/redis"
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

const (
	// Redis and our redis client library allow lua scripting. We use the following script to conditionally decrement
	// a key-value pair by a given amount. This allows us not only to do both actions in a single round-trip to the
	// database but provide a lock-free implementation for logic that would otherwise be unsafe in a concurrent environment.
	//
	// Consider if we make two calls: .Get() and .Decr() with a conditional that .Decr() should only be called if .Get()
	// returns a certain amount... what if the value had been modified by a separate client during the
	// conditional statement? Then we might decrement the value incorrectly.
	//
	// Notably lua scripting is less performent then normal Redis commands thus this library should not share a database
	// with a super high-performance requirement
	//
	// Notice the use of tonumber() in Lua because everything Redis takes in and spits out is a string
	luaGetAndDecr = `
		local key = KEYS[1]
		local amount = tonumber(ARGV[1])
		local count = tonumber(redis.call("get", key))

		if count >= amount then
			return redis.call("DECRBY", key, amount)
		else
			error("Insufficient tokens/")
		end
	`
)

// Bucket is a superset struct of redis.Client and will contain all the methods of redis.Client plus those defined below.
type Bucket struct {
	*redis.Client

	// the name of the bucket, used as the key in Redis from which the token value is stored
	Name string

	// the token value a bucket should hold when it is created, if the bucket already exists this does nothing
	capacity int
}

// Connect and verify a redis client connection while instantiating a bucket then create a key-value pair in the database
// if one does not already exist for the given name
func NewBucket(name string, capacity int, storageOptions *redis.Options) (*Bucket, error) {
	bucket := &Bucket{ redis.NewClient(storageOptions), name, capacity }

	// ensure our redis connection is valid
	_, err := bucket.Ping().Result()
	if err != nil {
		return nil, err
	}

	err = bucket.Create(name, capacity)

	return bucket, err
}

// bucket.Create will create a new bucket with the given parameters if one does not exist, if no bucket can be created it will return an error
func (bucket *Bucket) Create(name string, capacity int) error {
	// query redis and see if the key already exists
	_, err := bucket.Get(name).Result()

	// err == redis.Nil would indicate the key does not exist which is an acceptable error
	if err != nil && err != redis.Nil {
		return err
	}

	// if the error is for a key that does not exist create a key with a value of capacity
	if err == redis.Nil {
		// the last param 0 indicates the key will never expire
		if _, err = bucket.Set(name, capacity, 0).Result(); err != nil {
			return err
		}
	}

	// great we know the bucket exists and has been created if it doesn't already exist
	return nil
}

// Executes a lua script which decrements the token value by tokensDesired if tokensDesired >= the token value.
func (bucket *Bucket) Take(tokensDesired int) error {
	return bucket.Eval(luaGetAndDecr, []string{bucket.Name}, tokensDesired).Err()
}

// Increment the token value by a given amount
func (bucket *Bucket) Put(amount int) error {
	return bucket.IncrBy(bucket.Name, int64(amount)).Err()
}

// attempt on a 500ms interval to asynchronously call bucket.Take until timeout is exceeded
// returns a channel which will fire nil or error on completion
func (bucket *Bucket) Watch(tokensDesired int, timeout time.Duration) chan error {
	done := make(chan error)

	go func(tokensDesired int, timeout time.Duration, done chan error){
		// time.Ticker returns a channel which fires every time the duration provided is passed
		ticker := time.NewTicker(time.Millisecond * 500)
		defer ticker.Stop()

		for {
			select {

			// attempt to take the desiredTokens on every ticker event
			case <-ticker.C:
				if err := bucket.Take(tokensDesired); err == nil {
					done <- nil
					break
				}

			// return an error if our timeout has passed
			case <-time.After(timeout):
				done <- errors.New("Watch timeout.")
				break
			}
		}
	}(tokensDesired, timeout, done)

	return done
}