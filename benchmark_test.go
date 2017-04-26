package bucket_test

import (
	tb "github.com/b3ntly/bucket"
	"testing"
)

func BenchmarkBucket_Create(b *testing.B) {
	cases := []*BucketCase{
		&BucketCase{ constructor: tb.New, options: &tb.Options{} },
		&BucketCase{ constructor: tb.NewWithRedis, options: redisBucketOptions },
	}

	for _, test := range cases {
		b.Run("Bucket_Create_Benchmark", func(b *testing.B){
			for i := 0; i < b.N; i++ {
				test.options.Name = MockBucketName()
				test.options.Capacity= 10
				_, _ = test.constructor(test.options)
			}
		})

	}
}

func BenchmarkBucket_Take(b *testing.B) {
	cases := []*BucketCase{
		{ constructor: tb.New, options: &tb.Options{} },
		{ constructor: tb.NewWithRedis, options: redisBucketOptions },
	}

	for _, test := range cases {
		test.options.Name = MockBucketName()
		test.options.Capacity= b.N
		bucket, _ := test.constructor(test.options)

		b.Run("Bucket_Take_Benchmark", func(b *testing.B){
			for i := 0; i < b.N; i++ {
				_ = bucket.Take(1)
			}
		})
	}
}

func BenchmarkBucket_Put(b *testing.B) {
	cases := []*BucketCase{
		{ constructor: tb.New, options: &tb.Options{} },
		{ constructor: tb.NewWithRedis, options: redisBucketOptions },
	}

	for _, test := range cases {
		test.options.Name = MockBucketName()
		test.options.Capacity = 0
		bucket, _ := test.constructor(test.options)

		b.Run("Bucket_Take_Benchmark", func(b *testing.B){
			for i := 0; i < b.N; i++ {
				_ = bucket.Put(1)
			}
		})
	}
}
