package main

import (
	tb "github.com/b3ntly/distributed-token-bucket"
	storage "github.com/b3ntly/distributed-token-bucket/storage"
	"github.com/go-redis/redis"
	"fmt"
)

func main(){
	var err error

	storageOptions := &redis.Options{
		Addr:     "127.0.0.1:6379",
		PoolSize: 30,
	}

	store, err := storage.NewStorage("redis", storageOptions)
	// error == nil

	// you can create multiple buckets with the same storage instance to share client connections
	bucketOne, err := tb.NewBucket("bucket_one", 50, store)
	bucketTwo, err := tb.NewBucket("bucket_two", 50, store)
	bucketThree, err := tb.NewBucket("bucket_three", 50, store)

	err = bucketOne.Take(5)
	err = bucketTwo.Take(5)
	err = bucketThree.Take(5)

	fmt.Println(err)
}