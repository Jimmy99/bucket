package storage_test

import (
	"fmt"
	"testing"
	"sync/atomic"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	tb "github.com/b3ntly/bucket"
	"github.com/b3ntly/bucket/storage"
)

func MockRedisBucket(capacity int, redisOptions *redis.Options) (*tb.Bucket, error) {
	// create unique bucket names from a concurrently accessible index
	atomic.AddInt32(&bucketIndex, 1)
	name := fmt.Sprintf("bucket_%v", atomic.LoadInt32(&bucketIndex))

	return tb.NewWithRedis(&tb.Options{
		Name: name,
		Capacity: capacity,
		Storage: &storage.RedisStorage{
			Client: redis.NewClient(redisOptions),
		},

	})
}

func TestRedisStorage_Ping(t *testing.T) {
	asserts := assert.New(t)
	testClient := redis.NewClient(redisOptions)


	// redisStorage specific tests
	t.Run("NewBucket will contain an error if storage.Ping() fails", func(t *testing.T) {
		_, err := MockRedisBucket(10, brokenRedisOptions)
		asserts.Error(err, "brokenBucket test did not return an error for an invalid redis connection.")
	})

	t.Run("NewBucket should create a key in Redis", func(t *testing.T) {
		bucket, err := MockRedisBucket(10, redisOptions)
		asserts.Nil(err, "Failed to create bucket for tb.Create()")

		err = testClient.Get(bucket.Name).Err()
		asserts.Nil(err, "Bucket name for tb.Create() test not saved in redis.")
	})

	t.Run("bucket.Create will error if the key is already taken by a value that cannot be converted to an integer", func(t *testing.T) {
		err := testClient.Set("some_key", "some_value", 0).Err()
		asserts.Nil(err, "Incorrectly returned an error for client.Set().keyTaken")

		_, err = tb.New(&tb.Options{
			Capacity: 10,
			Name: "some_key",
			Storage: &storage.RedisStorage{
				Client: redis.NewClient(redisOptions),
			},
		})

		asserts.Error(err, "Failed to return an error for NewBucket().keyTaken")
	})


	t.Run("bucket.Create will error if the key is already taken and contains a value of 0.", func(t *testing.T) {
		err := testClient.Set("some_key2", 0, 0).Err()
		asserts.Nil(err, "Incorrectly returned an error for client.Set().keyTaken")

		_, err = tb.New(&tb.Options{
			Capacity: 10,
			Name: "some_key2",
			Storage: &storage.RedisStorage{
				Client: redis.NewClient(redisOptions),
			},
		})

		asserts.Error(err, "Failed to return an error for Newbucket().keyZero")
	})

	err := testClient.FlushDb().Err()
	asserts.Nil(err, "redist test db should flush")
}