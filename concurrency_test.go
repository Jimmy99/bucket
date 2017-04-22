package distributed_token_bucket_test

import (
	"testing"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

// concurrency is hard to test for and this library should be evaluated
// much more stringently then these tests provide for if used in production
func TestTokenBucketConcurrency(t *testing.T){
	asserts := assert.New(t)
	testClient := redis.NewClient(redisOptions)

	t.Run("bucket.Take is safe for basic concurrent access", func(t *testing.T) {
		bucket, err := MockBucket(10)
		asserts.Nil(err, "error 1 should be nil")

		iterations := make(chan int)
		done := make(chan bool)

		go func() {
			for {
				_, ok := <-iterations

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

		tokenCount, err := testClient.Get(bucket.Name).Int64()

		asserts.Nil(err, "error 3 should be nothing")
		assert.Equal(t, int64(0), tokenCount, "testBucket should now have 0 tokens")
	})
}