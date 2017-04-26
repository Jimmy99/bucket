package bucket_test

import (
	tb "github.com/b3ntly/bucket"
	"github.com/b3ntly/bucket/storage"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
	"fmt"
)

// INTEGRATION TESTING

var (
	redisOptions = &redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       5,  // use default DB
	}

	redisBucketOptions = &tb.Options{
		Storage: &storage.RedisStorage{
			Client: redis.NewClient(redisOptions),
		},
	}

	bucketIndex int32 = 0

	testClient = redis.NewClient(redisOptions)
)

type BucketCase struct {
	constructor func(options *tb.Options)(*tb.Bucket, error)
	options *tb.Options
}

// return a unique bucket name
func MockBucketName() string {
	// create unique bucket names from a concurrently accessible index
	atomic.AddInt32(&bucketIndex, 1)
	return fmt.Sprintf("bucket_%v", atomic.LoadInt32(&bucketIndex))
}


func MockStorage() []storage.Storage {
	redisStorage := &storage.RedisStorage{ Client: redis.NewClient(redisOptions) }
	memoryStorage := &storage.MemoryStorage{}
	return []storage.Storage{ redisStorage, memoryStorage }
}


func TestTokenBucket(t *testing.T) {
	var err error

	asserts := assert.New(t)
	asserts.Nil(err, "Should be able to create a broken redis storage instance")

	cases := []*BucketCase{
		&BucketCase{ constructor: tb.New, options: &tb.Options{} },
		&BucketCase{ constructor: tb.NewWithRedis, options: redisBucketOptions },
	}

	// provider agnostic tests which should be run against each provider
	for _, test := range cases {
		t.Run("Bucket should be countable", func(t *testing.T){
			test.options.Name = MockBucketName()
			test.options.Capacity= 10
			bucket, err := test.constructor(test.options)

			asserts.Nil(err, "Failed to create bucket for .Count")

			count, err := bucket.Count()
			asserts.Equal(test.options.Capacity, count, "count should be equal")
		})
		t.Run("Cannot take more then initialCapacity from testBucket", func(t *testing.T) {
			test.options.Name = MockBucketName()
			test.options.Capacity = 10
			bucket, err := test.constructor(test.options)
			asserts.Nil(err, "Failed to create bucket for initialCapacity test.")

			err = bucket.Take(11)
			asserts.Error(err, "Failed to return an error for initialCapacity test.")
		})

		t.Run("Can take more then initialCapacity if more then initial capacity is Put() in", func(t *testing.T) {
			test.options.Name = MockBucketName()
			test.options.Capacity = 10
			bucket, err := test.constructor(test.options)
			asserts.Nil(err, "Failed to create bucket for .Put() test")

			err = bucket.Put(1)
			asserts.Nil(err, ".Put() incorrectly returned an error")

			err = bucket.Take(11)
			asserts.Nil(err, ".Take() incorrectly returned an error after .Put()")
		})

		t.Run("bucket.Watch will return nil before timeout if enough tokens are put in", func(t *testing.T) {
			test.options.Name = MockBucketName()
			test.options.Capacity = 10
			bucket, err := test.constructor(test.options)

			asserts.Nil(err, "Incorrectly returned an error for bucket.Watch() test")

			// call bucket.Watch with a one minute timeout, this becomes a race condition but *should* never matter
			done := bucket.Watch(11, time.Second*10).Done()
			err = bucket.Put(1)
			asserts.Nil(err, "Incorrectly returned an error on bucket.Watch() test (2)")

			err = <-done

			asserts.Nil(err, "Incorrectly returned an error on bucket.Watch() test (3)")
		})

		t.Run("bucket.Watch will return an error if timeout is exceeded", func(t *testing.T) {
			test.options.Name = MockBucketName()
			test.options.Capacity = 10
			bucket, err := test.constructor(test.options)

			asserts.Nil(err, "Failed to create a bucket for bucket.Watch().timeout test")

			// call bucket.Watch with a one minute timeout, this becomes a race condition but *should* never matter
			done := bucket.Watch(11, time.Millisecond*1).Done()
			err = <-done
			asserts.Error(err, "Failed to return an error due to a timeout on bucket.Watch()")
		})

		t.Run("bucket.Count should count", func(t *testing.T){
			test.options.Name = MockBucketName()
			test.options.Capacity = 10
			bucket, err := test.constructor(test.options)

			asserts.Nil(err, "Failed to create a bucket for bucket.Fill.cancelable test")

			count, err := bucket.Count()
			asserts.Nil(err, "bucket.Count should not return an error for bucket.Fill")

			asserts.NotZero(count, "Count should be greater then 0")
		})

		t.Run("bucket.Fill can be canceled", func(t *testing.T){
			test.options.Name = MockBucketName()
			test.options.Capacity = 0
			bucket, err := test.constructor(test.options)
			asserts.Nil(err, "Failed to create a bucket for bucket.Fill.cancelable test")

			watchable := bucket.Fill(100, time.Second * 1)
			done := watchable.Done()
			watchable.Cancel <- nil
			err = <-done
			asserts.Nil(err, "bucket.Fill should cancel without an error")
		})

		t.Run("bucket.Fill actually fills", func(t *testing.T){
			test.options.Name = MockBucketName()
			test.options.Capacity = 100
			bucket, err := test.constructor(test.options)
			asserts.Nil(err, "Failed to create a bucket for bucket.Fill.cancelable test")

			watchable := bucket.Fill(100, time.Millisecond * 1)
			done := watchable.Done()
			time.Sleep(time.Millisecond * 5)

			watchable.Cancel <- nil
			<-done

			asserts.Nil(err, "bucket.Fill should cancel without an error")

			count, err := bucket.Count()
			asserts.Nil(err, "bucket.Count should not return an error for bucket.Fill")

			asserts.NotZero(count, "Count should be greater then 0")
		})

		t.Run("bucket.DynamicFill will increase the token value when a channel is sent to", func(t *testing.T){
			t.Run("bucket.Fill actually fills", func(t *testing.T){
				test.options.Name = MockBucketName()
				test.options.Capacity = 100
				bucket, err := test.constructor(test.options)
				asserts.Nil(err, "Failed to create a bucket for bucket.DynamicFill.cancelable test")

				signal := make(chan time.Time)
				watchable := bucket.DynamicFill(100, signal)
				done := watchable.Done()
				signal <- time.Now()

				watchable.Cancel <- nil
				<-done

				asserts.Nil(err, "bucket.Fill should cancel without an error")

				count, err := bucket.Count()
				asserts.Nil(err, "bucket.Count should not return an error for bucket.Fill")

				asserts.NotZero(count, "Count should be greater then 0")
			})
		})

		// I'm really only writing this one for the sweet test coverage karma, it's covered by a unit test in ./storage
		t.Run("bucket.TakeAll will return the current token value of a bucket then set it to zero", func(t *testing.T){
			expectedCount := 12
			test.options.Name = MockBucketName()
			test.options.Capacity = expectedCount
			bucket, err := test.constructor(test.options)
			asserts.Nil(err, "Should be able to create a bucket for store.TakeAll test")

			count, err := bucket.TakeAll()
			asserts.Nil(err, "bucket.TakeAll should not return an error")
			asserts.Equal(expectedCount, count, "count should equal expectedCount")

			finalCount, err := bucket.Count()
			asserts.Nil(err, "store.Count should not return an error")
			asserts.Equal(0, finalCount, "count should equal expectedCount")
		})
	}

	err = testClient.FlushDb().Err()
	asserts.Nil(err, "redist test db should flush")
}
