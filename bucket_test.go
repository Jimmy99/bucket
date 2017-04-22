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

	bucketIndex = int32(0)
)

func MockBucket(initialCapacity int) (*tb.Bucket, error) {
	atomic.AddInt32(&bucketIndex, 1)
	return tb.NewBucket(fmt.Sprintf("bucket_%v", atomic.LoadInt32(&bucketIndex)), initialCapacity, redisOptions)
}

func TestTokenBucket(t *testing.T) {
	asserts := assert.New(t)
	testClient := redis.NewClient(redisOptions)

	t.Run("bucket.Ping returns redis connection ping response", func(t *testing.T) {
		bucket, err := MockBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		_, err = bucket.Ping().Result()
		asserts.Nil(err, "error 2 should be nil")
	})

	t.Run("NewBucket will contain an error if a redis connection is invalid", func(t *testing.T) {
		_, err := tb.NewBucket("brokenBucket", 10, brokenRedisOptions)

		asserts.NotNil(err, "error should not be nil")
	})

	t.Run("NewBucket should create a key in Redis", func(t *testing.T) {
		bucket, err := MockBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		err = testClient.Get(bucket.Name).Err()
		asserts.Nil(err, "error 2 should be nothing")
	})

	t.Run("Cannot take more then initialCapacity from testBucket", func(t *testing.T) {
		bucket, err := MockBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		err = bucket.Take(11)
		asserts.NotNil(err, "error 2 should be nil")

		tokenCount, err := testClient.Get(bucket.Name).Int64()

		asserts.Nil(err, "error 3 should be nothing")
		assert.Equal(t, int64(10), tokenCount, "testBucket should still have 10 tokens")
	})

	t.Run("Can take more then initialCapacity if more then initial capacity is Put() in", func(t *testing.T) {
		bucket, err := MockBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		err = bucket.Put(1)
		asserts.Nil(err, "error 2 should be nil")

		err = bucket.Take(11)
		asserts.Nil(err, "error 3 should be nothing")

		tokenCount, err := testClient.Get(bucket.Name).Int64()

		asserts.Nil(err, "error 3 should be nothing")
		assert.Equal(t, int64(0), tokenCount, "testBucket should still have 10 tokens")
	})

	t.Run("bucket.Watch will return nil before timeout if enough tokens is put in", func(t *testing.T){
		bucket, err := MockBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		// call bucket.Watch with a one minute timeout, this becomes a race condition but *should* never matter
		done := bucket.Watch(11, time.Minute * 1)
		err = bucket.Put(1)
		asserts.Nil(err, "error 2 should be nil")

		err = <- done

		asserts.Nil(err, "watch should not concurrently return an error")
	})

	t.Run("bucket.Watch will return an error if timeout is exceeded", func(t *testing.T){
		bucket, err := MockBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		// call bucket.Watch with a one minute timeout, this becomes a race condition but *should* never matter
		done := bucket.Watch(11, time.Millisecond * 1)
		err = <- done
		asserts.NotNil(err, "error should not be nil")
	})

	err := testClient.FlushDb().Err()
	asserts.Nil(err, "redist test db should flush")
}
