package distributed_token_bucket_test

import (
	tb "github.com/b3ntly/distributed-token-bucket"
	"github.com/stretchr/testify/assert"
	"github.com/go-redis/redis"
	"sync/atomic"
	"testing"
	"time"
	"fmt"
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

func MockBucket(initialCapacity int) (*tb.Bucket, error) {
	atomic.AddInt32(&bucketIndex, 1)
	name := fmt.Sprintf("bucket_%v", atomic.LoadInt32(&bucketIndex))
	return tb.NewBucket(name, initialCapacity, redisOptions)
}

func TestTokenBucket(t *testing.T) {
	asserts := assert.New(t)

	t.Run("NewBucket will contain an error if a redis connection is invalid", func(t *testing.T) {
		_, err := tb.NewBucket("brokenBucket", 10, brokenRedisOptions)

		asserts.Error(err, "brokenBucket test did not return an error for an invalid redis connection.")
	})

	t.Run("NewBucket should create a key in Redis", func(t *testing.T) {
		bucket, err := MockBucket(10)
		asserts.Nil(err, "Failed to create bucket for tb.Create()")

		err = testClient.Get(bucket.Name).Err()
		asserts.Nil(err, "Bucket name for tb.Create() test not saved in redis.")
	})

	t.Run("Cannot take more then initialCapacity from testBucket", func(t *testing.T) {
		bucket, err := MockBucket(10)
		asserts.Nil(err, "Failed to create bucket for initialCapacity test.")

		err = bucket.Take(11)
		asserts.Error(err, "Failed to return an error for initialCapacity test.")

		tokenCount, err := testClient.Get(bucket.Name).Int64()
		asserts.Nil(err, "Failed to create bucket for initial capacity test (2).")

		assert.Equal(t, int64(10), tokenCount, "testBucket should still have 10 tokens")
	})

	t.Run("Can take more then initialCapacity if more then initial capacity is Put() in", func(t *testing.T) {
		bucket, err := MockBucket(10)
		asserts.Nil(err, "Failed to create bucket for .Put() test")

		err = bucket.Put(1)
		asserts.Nil(err, ".Put() incorrectly returned an error")

		err = bucket.Take(11)
		asserts.Nil(err, ".Take() incorrectly returned an error after .Put()")

		tokenCount, err := testClient.Get(bucket.Name).Int64()
		asserts.Nil(err, "Failed to create a bucket for .Put() test (2)")

		assert.Equal(t, int64(0), tokenCount, "testBucket should still have 10 tokens")
	})

	t.Run("bucket.Watch will return nil before timeout if enough tokens is put in", func(t *testing.T){
		bucket, err := MockBucket(10)
		asserts.Nil(err, "Incorrectly returned an error for bucket.Watch() test")

		// call bucket.Watch with a one minute timeout, this becomes a race condition but *should* never matter
		done := bucket.Watch(11, time.Second * 10)
		err = bucket.Put(1)
		asserts.Nil(err, "Incorrectly returned an error on bucket.Watch() test (2)")

		err = <- done

		asserts.Nil(err, "Incorrectly returned an error on bucket.Watch() test (3)")
	})

	t.Run("bucket.Watch will return an error if timeout is exceeded", func(t *testing.T){
		bucket, err := MockBucket(10)
		asserts.Nil(err, "Failed to create a bucket for bucket.Watch().timeout test")

		// call bucket.Watch with a one minute timeout, this becomes a race condition but *should* never matter
		done := bucket.Watch(11, time.Millisecond * 1)
		err = <- done
		asserts.Error(err, "Failed to return an error due to a timeout on bucket.Watch()")
	})

	t.Run("bucket.Create will error if the key is already taken by a value that cannot be converted to an integer", func(t *testing.T){
		err := testClient.Set("some_key", "some_value", 0).Err()
		asserts.Nil(err, "Incorrectly returned an error for client.Set().keyTaken")

		_, err = tb.NewBucket("some_key", 10, redisOptions)
		asserts.Error(err, "Failed to return an error for NewBucket().keyTaken")
	})

	t.Run("bucket.Create will error if the key is already taken and contains a value of 0.", func(t *testing.T){
		err := testClient.Set("some_key2", 0, 0).Err()
		asserts.Nil(err, "Incorrectly returned an error for client.Set().keyTaken")

		_, err = tb.NewBucket("some_key2", 10, redisOptions)
		asserts.Error(err, "Failed to return an error for Newbucket().keyZero")
	})

	err := testClient.FlushDb().Err()
	asserts.Nil(err, "redist test db should flush")
}
