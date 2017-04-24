package distributed_token_bucket_test

import (
	"fmt"
	tb "github.com/b3ntly/distributed-token-bucket"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

var (
	redisOptions = &redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       5,  // use default DB
	}

	brokenRedisOptions = &redis.Options{
		Addr: "127.0.0.1:8080",
	}

	bucketIndex int32 = 0

	testClient = redis.NewClient(redisOptions)
)


func MockBucket(initialCapacity int, storage tb.IStorage) (*tb.Bucket, error) {
	// create unique bucket names from a concurrently accessible index
	atomic.AddInt32(&bucketIndex, 1)
	name := fmt.Sprintf("bucket_%v", atomic.LoadInt32(&bucketIndex))

	return tb.NewBucket(name, initialCapacity, storage)
}

func MockStorage() []tb.IStorage {
	redisStorage, _ := tb.NewStorage("redis", redisOptions)
	memoryStorage, _ := tb.NewStorage("memory", nil)
	return []tb.IStorage{ redisStorage, memoryStorage }
}


func TestTokenBucket(t *testing.T) {
	var err error

	asserts := assert.New(t)

	brokenRedisStorage, err := tb.NewStorage("redis", brokenRedisOptions)
	asserts.Nil(err, "Should be able to create a broken redis storage instance")

	// provider agnostic tests which should be run against each provider
	for _, storage := range MockStorage() {
		t.Run("Cannot take more then initialCapacity from testBucket", func(t *testing.T) {
			bucket, err := MockBucket(10, storage)
			asserts.Nil(err, "Failed to create bucket for initialCapacity test.")

			err = bucket.Take(11)
			asserts.Error(err, "Failed to return an error for initialCapacity test.")
		})

		t.Run("Can take more then initialCapacity if more then initial capacity is Put() in", func(t *testing.T) {
			bucket, err := MockBucket(10, storage)
			asserts.Nil(err, "Failed to create bucket for .Put() test")

			err = bucket.Put(1)
			asserts.Nil(err, ".Put() incorrectly returned an error")

			err = bucket.Take(11)
			asserts.Nil(err, ".Take() incorrectly returned an error after .Put()")
		})

		t.Run("bucket.Watch will return nil before timeout if enough tokens are put in", func(t *testing.T) {
			bucket, err := MockBucket(10, storage)
			asserts.Nil(err, "Incorrectly returned an error for bucket.Watch() test")

			// call bucket.Watch with a one minute timeout, this becomes a race condition but *should* never matter
			done := bucket.Watch(11, time.Second*10)
			err = bucket.Put(1)
			asserts.Nil(err, "Incorrectly returned an error on bucket.Watch() test (2)")

			err = <-done

			asserts.Nil(err, "Incorrectly returned an error on bucket.Watch() test (3)")
		})

		t.Run("bucket.Watch will return an error if timeout is exceeded", func(t *testing.T) {
			bucket, err := MockBucket(10, storage)
			asserts.Nil(err, "Failed to create a bucket for bucket.Watch().timeout test")

			// call bucket.Watch with a one minute timeout, this becomes a race condition but *should* never matter
			done := bucket.Watch(11, time.Millisecond*1)
			err = <-done
			asserts.Error(err, "Failed to return an error due to a timeout on bucket.Watch()")
		})


	}

	// redisStorage specific tests
	t.Run("NewBucket will contain an error if storage.Ping() fails", func(t *testing.T) {
		_, err := MockBucket(10, brokenRedisStorage)
		asserts.Error(err, "brokenBucket test did not return an error for an invalid redis connection.")
	})

	t.Run("NewBucket should create a key in Redis", func(t *testing.T) {
		bucket, err := MockBucket(10, MockStorage()[0])
		asserts.Nil(err, "Failed to create bucket for tb.Create()")

		err = testClient.Get(bucket.Name).Err()
		asserts.Nil(err, "Bucket name for tb.Create() test not saved in redis.")
	})

	t.Run("bucket.Create will error if the key is already taken by a value that cannot be converted to an integer", func(t *testing.T) {
		err := testClient.Set("some_key", "some_value", 0).Err()
		asserts.Nil(err, "Incorrectly returned an error for client.Set().keyTaken")

		_, err = tb.NewBucket("some_key", 10, brokenRedisStorage)
		asserts.Error(err, "Failed to return an error for NewBucket().keyTaken")
	})


	t.Run("bucket.Create will error if the key is already taken and contains a value of 0.", func(t *testing.T) {
		err := testClient.Set("some_key2", 0, 0).Err()
		asserts.Nil(err, "Incorrectly returned an error for client.Set().keyTaken")

		_, err = tb.NewBucket("some_key2", 10, MockStorage()[0])
		asserts.Error(err, "Failed to return an error for Newbucket().keyZero")
	})

	err = testClient.FlushDb().Err()
	asserts.Nil(err, "redist test db should flush")
}
