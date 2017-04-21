package distributed_token_bucket

import (
	"fmt"
	"sync/atomic"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/go-redis/redis"
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

func buildBucket(initialCapacity int) (*Bucket, error) {
	atomic.AddInt32(&bucketIndex, 1)
	return NewBucket(fmt.Sprintf("bucket_%s", atomic.LoadInt32(&bucketIndex)), initialCapacity, redisOptions)
}

func TestTokenBucket(t *testing.T) {
	asserts := assert.New(t)
	testClient := redis.NewClient(redisOptions)

	t.Run("bucket.Ping returns redis connection ping response", func(t *testing.T){
		bucket, err := buildBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		_, err = bucket.Ping()
		asserts.Nil(err, "error 2 should be nil")
	})

	t.Run("NewBucket will contain an error if a redis connection is invalid", func(t *testing.T){
		_, err := NewBucket("brokenBucket", 10, brokenRedisOptions)

		asserts.NotNil(err, "error should not be nil")
	})

	t.Run("storage.Create will return an error from an invalid redis connection", func(t *testing.T){
		brokenStorage := NewStorage(brokenRedisOptions)
		err := brokenStorage.Create("broken bucket", 10)
		asserts.NotNil(err, "error should not be nil")
	})

	t.Run("NewBucket should create a key in Redis", func(t *testing.T){
		bucket, err := buildBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		err = testClient.Get(bucket.name).Err()
		asserts.Nil(err, "error 2 should be nothing")
	})

	t.Run("Cannot take more then initialCapacity from testBucket", func (t *testing.T){
		bucket, err := buildBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		err = bucket.Take(11)
		asserts.NotNil(err, "error 2 should be nil")

		tokenCount, err := testClient.Get(bucket.name).Int64()

		asserts.Nil(err, "error 3 should be nothing")
		assert.Equal(t, int64(10), tokenCount, "testBucket should still have 10 tokens")
	})

	t.Run("Can take more then initialCapacity if more then initial capacity is Put() in", func(t *testing.T){
		bucket, err := buildBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		err = bucket.Put(1)
		asserts.Nil(err, "error 2 should be nil")

		err = bucket.Take(11)
		asserts.Nil(err, "error 3 should be nothing")

		tokenCount, err := testClient.Get(bucket.name).Int64()

		asserts.Nil(err, "error 3 should be nothing")
		assert.Equal(t, int64(0), tokenCount, "testBucket should still have 10 tokens")
	})

	t.Run("bucket.Take is safe for basic concurrent access", func(t *testing.T){
		bucket, err := buildBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		iterations := make(chan int)
		done := make(chan bool)

		go func(){
			for {
				_, ok := <- iterations

				if !ok {
					done <- true
					return
				}

				err := bucket.Take(2)
				asserts.Nil(err, "error 2 should be nothing")
			}
		}()

		for i := 0; i < 5; i++ {
			iterations <- i
		}

		// wait for all go routines to finish
		close(iterations)
		<-done

		tokenCount, err := testClient.Get(bucket.name).Int64()

		asserts.Nil(err, "error 3 should be nothing")
		assert.Equal(t, int64(0), tokenCount, "testBucket should now have 0 tokens")
	})

	err := testClient.FlushDb().Err()
	asserts.Nil(err, "redist test db should flush")
}