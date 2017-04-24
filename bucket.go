package bucket

import (
	"github.com/b3ntly/bucket/storage"
	"errors"
	"time"
)

/**
 * bucket.go provides the operations for a bucket.
 *
 * Some methods: Create, Take, Put, Count, Watch, Fill, Ping
 *
 * Originally this library was meant to implement a token bucket algorithm but I realized it could be
 * made much more abstract then that. From this library you can easily implement a distributed token bucket algorithm
 * (via redis to share storage between nodes) or local rate-limiting, etc.
 *
 * Storage is a simple interface that you can understand by going through the ./storage directory. It should be
 * trivial to add other persistence engines like memcached as long as it can fulfill the interface. Storage is designed
 * to be as dumb as possible such that instances may be shared between buckets.
 *
 * Generally speaking buckets should also be 'cheap', I imagine implementing it on a per user-session basis for
 * rate-limiting.
 *
 * For Redis the bucket name is a key in the database. For memory it is a key in a map that is protected by a RWmutex.
 *
 * Bucket behavior will vary between storage implementations so you should check out each provider. For example RedisStorage
 * will offer some basic protections such as not overriding string-based values for new buckets. It will also prevent creating
 * new buckets if the key already exists with a value of 0, this is in case a key was expected not to be there but actually
 * was. It's not perfect but hopes to be intuitive.
 */

type (
	Bucket struct {
		storage storage.IStorage

		// the name of the bucket, may have implications for certain storage providers
		Name string

		// the token value a bucket should hold when it is created, if the bucket already exists this does nothing
		capacity int
	}
)

// Instantiate the bucket with a given storage object. Test the validity of the storage with storage.Ping().
//
// If a name already exists its storage will be shared with the bucket that created it.
//
// The storage provider might reject a bucket name (and return an error) if the value already assigned to a bucket is
// unexpected, for example if a name has a string value in redis (which may indicate it is reserved for something else
// in the database). Redis will also reject sharing a bucket name whose value is 0, this was a personal choice because
// I found myself incorrectly using bucket names I thought did not exist but actually did (from leftover tests).
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

func (bucket *Bucket) Create(name string, capacity int) error {
	return bucket.storage.Create(name, capacity)
}

// Decrement the token value of a bucket if the number of tokens in the bucket is >= tokensDesired. It will return
// an error if not enough tokens exist.
func (bucket *Bucket) Take(tokensDesired int) error {
	return bucket.storage.Take(bucket.Name, tokensDesired)
}

// Increment the token value by a given amount.
func (bucket *Bucket) Put(amount int) error {
	return bucket.storage.Put(bucket.Name, amount)
}

// Return an integer count of a bucket's token value
func (bucket *Bucket) Count() (int, error) {
	return bucket.storage.Count(bucket.Name)
}

// Attempt on a 500ms interval to call bucket.Take with a nil response. It returns an instance of Watchable from which
// the polling can be cancelled and errors or nil may be received. See ./examples/watchable.go to get an idea of how it works.
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

// Start a ticker that will periodically put tokens into the bucket at a given rate on the defined interval. Returns
// a Watchable object identical to bucket.Watch, and thus may be canceled and observed.
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