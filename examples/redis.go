package main

import (
	"github.com/b3ntly/bucket"
	"github.com/b3ntly/bucket/storage"
	"github.com/go-redis/redis"
	"fmt"
)

func main(){
	storageOptions := &redis.Options{
		Addr:     "127.0.0.1:6379",
		PoolSize: 30,
	}

	store, err := storage.NewStorage("redis", storageOptions)

	if err != nil {
		// handle error
	}

	b, err := bucket.NewBucket("some_bucket", 10, store)

	if err != nil {
		// handle error
	}

	fmt.Println(b.Name)
}