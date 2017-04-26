package bucket

import (
	"github.com/b3ntly/bucket/storage"
	"errors"
	"time"
	"github.com/go-redis/redis"
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
		storage storage.Storage

		// the name of the bucket, may have implications for certain storage providers
		Name string

		// the token value a bucket should hold when it is created, if the bucket already exists this does nothing
		capacity int
	}

	Options struct {
		Storage storage.Storage
		Name string
		Capacity int
	}
)

var (
	DefaultMemoryStore = &storage.MemoryStorage{}

	// Options which use redis as the storage back-end, defaults to the default redis options
	DefaultRedisStore = &storage.RedisStorage{ Client: redis.NewClient(&redis.Options{ Addr: ":6379" }) }
)


// initialize options with defaults
func (opts *Options) init(store storage.Storage) *Options {
	if opts.Storage == nil {
		opts.Storage = store
	}

	return opts
}

// Instantiate the bucket with a given storage object. Test the validity of the storage with storage.Ping().
//
// If a name already exists its storage will be shared with the bucket that created it.
//
// The storage provider might reject a bucket name (and return an error) if the value already assigned to a bucket is
// unexpected, for example if a name has a string value in redis (which may indicate it is reserved for something else
// in the database). Redis will also reject sharing a bucket name whose value is 0, this was a personal choice because
// I found myself incorrectly using bucket names I thought did not exist but actually did (from leftover tests).
func New(options *Options) (*Bucket, error) {
	return create(options.init(DefaultMemoryStore))
}

// Create a bucket with Redis storage
func NewWithRedis(options *Options) (*Bucket, error){
	return create(options.init(DefaultRedisStore))
}

func create(options *Options) (*Bucket, error){
	bucket := &Bucket{Name: options.Name, capacity: options.Capacity, storage: options.Storage}

	// ensure our redis connection is valid
	err := bucket.storage.Ping()
	if err != nil {
		return nil, err
	}

	err = bucket.storage.Create(bucket.Name, bucket.capacity)
	return bucket, err
}

// Decrement the token value of a bucket if the number of tokens in the bucket is >= tokensDesired. It will return
// an error if not enough tokens exist.
func (bucket *Bucket) Take(tokensDesired int) error {
	return bucket.storage.Take(bucket.Name, tokensDesired)
}

// returns a conditional amount of tokens representing all the tokens
func (bucket *Bucket) TakeAll() (int, error){
	return bucket.storage.TakeAll(bucket.Name)
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

	go func(bucket *Bucket, watchable *Watchable, timeout <-chan time.Time, tokens int) {
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
	}(bucket, watchable, timeout, tokens)

	return watchable
}

// Start a ticker that will periodically set the token value to a given rate on the defined interval. Returns
// a Watchable object identical to bucket.Watch, and thus may be canceled and observed.
func (bucket *Bucket) Fill(rate int, interval time.Duration) *Watchable {
	watchable := NewWatchable()

	go func(bucket *Bucket, watchable *Watchable, rate int, interval time.Duration){
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// ensure we our not filling past capacity
		if rate > bucket.capacity {
			rate = bucket.capacity
		}

		for {
			select {
			case <- ticker.C:

				err := bucket.storage.Set(bucket.Name, rate)

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
	}(bucket, watchable, rate, interval)

	return watchable
}

// Dynamic fill fills the bucket every time it reads from interval
func (bucket *Bucket) DynamicFill(rate int, interval chan time.Time) *Watchable {
	watchable := NewWatchable()

	go func(bucket *Bucket, watchable *Watchable, rate int, interval chan time.Time){

		for {
			if rate > bucket.capacity {
				rate = bucket.capacity
			}

			select {
			case <- interval:


				err := bucket.storage.Set(bucket.Name, rate)

				// our program may rely on our bucket to be refilled reliably, so we should fail hard on error
				if err != nil {
					watchable.Failed <- err
					return
				}

			case err := <- watchable.Cancel:
				watchable.Failed <- err
				return
			}
		}
	}(bucket, watchable, rate, interval)

	return watchable
}