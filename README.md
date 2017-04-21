## Distributed Token Bucket with Redis and Golang

[example](./examples/basic.go)

```golang
package main

import (
	"log"
	"github.com/go-redis/redis"
	tb "github.com/b3ntly/distributed-token-bucket"
)

func main(){
	storageOptions := &redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	bucket, err := tb.NewBucket("test", 10, storageOptions)

	log.Println("Bucket has 10 tokens.")

	if err != nil {
		log.Println(err)
		return
	}

	err = bucket.Take(10)

	log.Println("Took 10 tokens.")

	if err != nil {
		log.Println(err)
		return
	}

	err = bucket.Take(10)

	if err != nil && err.Error() != "Insufficient tokens." {
		log.Println(err)
		return
	} else {
		log.Println("Insufficient tokens.")
	}

	err = bucket.Put(10)

	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Put 10 tokens")

	err = bucket.Take(10)

	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Took 10 tokens")
}
```