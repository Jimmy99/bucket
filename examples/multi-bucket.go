package main

import (
	"github.com/b3ntly/bucket"
	"github.com/b3ntly/bucket/storage"
	"github.com/go-redis/redis"
	"fmt"
)

func main(){
	var err error

	// with custom redis options
	store := &storage.RedisStorage{
		Client: redis.NewClient(&redis.Options{
			Addr: ":6379",
			DB: 5,
			PoolSize: 30,
		}),
	}

	b, err := bucket.New(&bucket.Options{
		Capacity: 10,
		Name: "My redis bucket with custom config 1",
		Storage: store,
	})

	b2, err := bucket.New(&bucket.Options{
		Capacity: 10,
		Name: "My redis bucket with custom config 2",
		Storage: store,
	})

	b3, err := bucket.New(&bucket.Options{
		Capacity: 10,
		Name: "My redis bucket with custom config 3",
		Storage: store,
	})

	err = b.Take(5)
	err = b2.Take(5)
	err = b3.Take(5)

	fmt.Println(err)
}