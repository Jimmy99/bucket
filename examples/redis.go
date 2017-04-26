package main

import (
	"github.com/b3ntly/bucket"
	"github.com/b3ntly/bucket/storage"
	"fmt"
	"github.com/go-redis/redis"
)

func main(){
	b, err := bucket.NewWithRedis(&bucket.Options{
		Capacity: 10,
		Name: "My redis bucket with default config",
	})

	if err != nil {
		fmt.Println(1, err)
		return
	}

	fmt.Println(b.Name)

	// with custom redis options
	store := &storage.RedisStorage{
		Client: redis.NewClient(&redis.Options{
			Addr: ":6379",
			DB: 5,
			PoolSize: 30,
		}),
	}

	b2, err := bucket.New(&bucket.Options{
		Capacity: 10,
		Name: "My redis bucket with custom config",
		Storage: store,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(b2.Name)
}