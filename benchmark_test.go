package distributed_token_bucket_test

import(
	tb "github.com/b3ntly/distributed-token-bucket"
	"github.com/go-redis/redis"
	"testing"
	"strconv"
)

var options = &redis.Options{
	Addr:     "127.0.0.1:6379",
	Password: "", // no password set
	DB:       5,  // use default DB
}

func BenchmarkBucket_Create(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = tb.NewBucket(strconv.Itoa(i), i, options)
	}
}

func BenchmarkBucket_Take(b *testing.B) {
	bucket, _ := MockBucket(b.N)
	for i := 0; i < b.N; i++ {
		_ = bucket.Take(1)
	}
}

func BenchmarkBucket_Put(b *testing.B) {
	bucket, _ := MockBucket(0)
	for i := 0; i < b.N; i++ {
		_ = bucket.Put(1)
	}
}