package distributed_token_bucket

import (
	"github.com/b3ntly/distributed-token-bucket/storage"
	"errors"
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

type (
	Bucket struct {
		storage storage.IStorage

		// the name of the bucket, used as the key in Redis for which the token value is stored
		Name string

		// the token value a bucket should hold when it is created, if the bucket already exists this does nothing
		capacity int
	}
)

// Connect and verify a redis client connection while instantiating a bucket then create a key-value pair in the database if one does not already exist for the given name
func NewBucket(name string, capacity int, storage storage.IStorage) (*Bucket, error) {
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

// return an integer count of a bucket's token value
func (bucket *Bucket) Count() (int, error) {
	return bucket.storage.Count(bucket.Name)
}

// attempt on a 500ms interval to asynchronously call bucket.Take until timeout is exceeded
// returns a channel which will fire nil or error on completion
func (bucket *Bucket) Watch(tokens int, duration time.Duration) *Watchable {
	watchable := NewWatchable()
	timeout := time.After(duration)

	go func() {
		// time.Ticker returns a channel which fires every time the duration provided is passed
		ticker := time.NewTicker(time.Millisecond * 500)
		defer ticker.Stop()

		for {
			select {

			// attempt to take the desiredTokens on every ticker event, expect an error though and continue the loop
			case <-ticker.C:
				// we expect an error here so we do not send into watchable.Failed
				if err := bucket.storage.Take(bucket.Name, tokens); err == nil {
					watchable.Success <- nil
					return
				}

			// cancel the watchable on timeout
			case <- timeout:
				watchable.Failed <- errors.New("Timeout.")
				return

			// on cancel end the ticker loop, possibly with an error
			case err := <-watchable.Cancel:
				watchable.Failed <- err
				return
			}
		}
	}()

	return watchable
}

// start a ticker that will periodically put tokens into the bucket at a given rate on a defined interval
func (bucket *Bucket) Fill(rate int, interval time.Duration) *Watchable {
	watchable := NewWatchable()

	go func(){
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {

			case <- ticker.C:
				count, err := bucket.storage.Count(bucket.Name)

				// our program may rely on our bucket to be refilled reliably, so we should fail hard on error
				if err != nil {
					watchable.Failed <- err
					return
				}

				err = bucket.storage.Put(bucket.Name, rate - count)

				// our program may rely on our bucket to be refilled reliably, so we should fail hard on error
				if err != nil {
					watchable.Failed <- err
					return
				}

			// on cancel end the ticker loop, possibly with an error
			case err := <-watchable.Cancel:
				watchable.Failed <- err
				return
			}
		}
	}()

	return watchable
}