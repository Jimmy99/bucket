package main

import (
	tb "github.com/b3ntly/distributed-token-bucket"
	"github.com/go-redis/redis"
	"fmt"
)

func main(){
	var err error

	storageOptions := &redis.Options{
		Addr:     "127.0.0.1:6379",
		PoolSize: 30,
	}

	storage, err := tb.NewStorage("redis", storageOptions)
	// error == nil

	// you can create multiple buckets with the same storage instance to share client connections
	bucketOne, err := tb.NewBucket("bucket_one", 50, storage)
	bucketTwo, err := tb.NewBucket("bucket_two", 50, storage)
	bucketThree, err := tb.NewBucket("bucket_three", 50, storage)

	err = bucketOne.Take(5)
	err = bucketTwo.Take(5)
	err = bucketThree.Take(5)

	fmt.Println(err)
}