package storage_test

import (
	"testing"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/b3ntly/bucket/storage"
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

	bucketIndex int32 = 0
)

func MockStorage() ([]storage.IStorage, error) {
	redisStorage, err := storage.NewStorage("redis", redisOptions)
	memoryStorage, err := storage.NewStorage("memory", nil)
	return []storage.IStorage{ redisStorage, memoryStorage }, err
}

func TestTokenBucket(t *testing.T) {
	var err error

	asserts := assert.New(t)
	testClient := redis.NewClient(redisOptions)

	stores, err := MockStorage()
	 asserts.Nil(err, "NewStorage does not create an error.")

	// provider agnostic tests which should be run against each provider
	for _, store := range stores {
		// create, take, put, count
		t.Run("store.Create does not return an error", func(t *testing.T){
			err := store.Create("some store", 0)
			asserts.Nil(err, "store.Create should not return an error")
		})

		t.Run("store.Put shound not return an error", func(t *testing.T){
			bucket, err := MockBucket(0, store)
			asserts.Nil(err, "Should be able to create a bucket for store.Put test")

			err = store.Put(bucket.Name, 1)
			asserts.Nil(err, "store.Put should not return an error")
		})

		t.Run("store.Take shound not return an error", func(t *testing.T){
			bucket, err := MockBucket(1, store)
			asserts.Nil(err, "Should be able to create a bucket for store.Take test")

			err = store.Take(bucket.Name, 1)
			asserts.Nil(err, "store.Take should not return an error")
		})

		t.Run("store.Count shound not return an error", func(t *testing.T){
			expected := 1

			bucket, err := MockBucket(expected, store)
			asserts.Nil(err, "Should be able to create a bucket for store.Count test")

			val, err := store.Count(bucket.Name, )
			asserts.Nil(err, "store.Take should not return an error")
			asserts.Equal(expected, val, "value == expected")
		})
	}

	err = testClient.FlushDb().Err()
	asserts.Nil(err, "redist test db should flush")
}