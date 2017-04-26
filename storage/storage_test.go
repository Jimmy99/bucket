package storage_test

import (
	"testing"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/b3ntly/bucket/storage"
	tb "github.com/b3ntly/bucket"
	"sync/atomic"
	"fmt"
)


// UNIT TESTING

var (
	redisOptions = &redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       5,  // use default DB
	}

	brokenRedisOptions = &redis.Options{
		Addr: "127.0.0.1:8080",
	}

	memoryBucketOptions = &tb.Options{}

	redisBucketOptions = &tb.Options{
		Storage: &storage.RedisStorage{
			Client: redis.NewClient(redisOptions),
		},
	}

	bucketIndex int32 = 0
)

func MockBucketName() string {
	// create unique bucket names from a concurrently accessible index
	atomic.AddInt32(&bucketIndex, 1)
	return fmt.Sprintf("bucket_%v", atomic.LoadInt32(&bucketIndex))
}

func MockStorage() []*tb.Options {
	return []*tb.Options{ memoryBucketOptions, redisBucketOptions }
}

func TestTokenBucket(t *testing.T) {
	var err error

	asserts := assert.New(t)
	testClient := redis.NewClient(redisOptions)

	// provider agnostic tests which should be run against each provider
	for _, options := range MockStorage() {


		t.Run("store.Put shound not return an error", func(t *testing.T){
			bucket, err := tb.New(options)
			asserts.Nil(err, "Should be able to create a bucket for store.Put test")

			err = options.Storage.Put(bucket.Name, 1)
			asserts.Nil(err, "store.Put should not return an error")
		})

		t.Run("store.Take shound not return an error", func(t *testing.T){
			options.Capacity = 1
			bucket, err := tb.New(options)
			asserts.Nil(err, "Should be able to create a bucket for store.Take test")

			err = options.Storage.Take(bucket.Name, 1)
			asserts.Nil(err, "store.Take should not return an error")
		})

		t.Run("store.Count shound not return an error", func(t *testing.T){
			expected := 1
			options.Capacity = 1
			options.Name = MockBucketName()
			bucket, err := tb.New(options)
			asserts.Nil(err, "Should be able to create a bucket for store.Count test")

			val, err := options.Storage.Count(bucket.Name)
			asserts.Nil(err, "store.Take should not return an error")
			asserts.Equal(expected, val, "value == expected")
		})

		t.Run("store.Set works", func(t *testing.T){
			bucket, err := tb.New(options)
			asserts.Nil(err, "Should be able to create a bucket for store.Count test")

			err = options.Storage.Set(bucket.Name, 10)
			asserts.Nil(err, "Set should return nil")

			val, err := options.Storage.Count(bucket.Name, )
			asserts.Nil(err, "store.Take should not return an error")
			asserts.Equal(10, val, "value == expected")
		})

		t.Run("store.TakeAll returns the token value and sets the token value to zero", func(t *testing.T){
			expectedCount := 12
			options.Capacity = expectedCount
			options.Name = MockBucketName()
			bucket, err := tb.New(options)
			asserts.Nil(err, "Should be able to create a bucket for store.TakeAll test")

			count, err := options.Storage.TakeAll(bucket.Name)
			asserts.Nil(err, "store.TakeAll should not return an error")
			asserts.Equal(expectedCount, count, "count should equal expectedCount")

			finalCount, err := options.Storage.Count(bucket.Name)
			asserts.Nil(err, "store.Count should not return an error")
			asserts.Equal(0, finalCount, "count should equal expectedCount")
		})
	}

	err = testClient.FlushDb().Err()
	asserts.Nil(err, "redist test db should flush")
}