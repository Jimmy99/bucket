package distributed_token_bucket_test

import (
	tb "github.com/b3ntly/distributed-token-bucket"
	"strconv"
	"testing"
)

func BenchmarkBucket_Create(b *testing.B) {
	for _, storage := range MockStorage() {
		b.Run("Bucket_Create_Benchmark", func(b *testing.B){
			for i := 0; i < b.N; i++ {
				_, _ = tb.NewBucket(strconv.Itoa(i), i, storage)
			}
		})

	}
}

func BenchmarkBucket_Take(b *testing.B) {
	for _, storage := range MockStorage() {
		bucket, _ := MockBucket(b.N, storage)

		b.Run("Bucket_Take_Benchmark", func(b *testing.B){
			for i := 0; i < b.N; i++ {
				_ = bucket.Take(1)
			}
		})
	}
}

func BenchmarkBucket_Put(b *testing.B) {
	for _, storage := range MockStorage() {
		bucket, _ := MockBucket(b.N, storage)

		b.Run("Bucket_Take_Benchmark", func(b *testing.B){
			for i := 0; i < b.N; i++ {
				_ = bucket.Put(1)
			}
		})
	}
}
