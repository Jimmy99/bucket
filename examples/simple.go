package main

import (
	tb "github.com/b3ntly/distributed-token-bucket"
	"time"
	"fmt"
)

func main(){
	storage, err := tb.NewStorage("memory", nil)
	// error == nil

	// initialize a bucket with 5 tokens
	bucket, err := tb.NewBucket("simple_bucket", 5, storage)

	// take 5 tokens
	err = bucket.Take(5)
	// error == nil

	// try to take 5 tokens, this will return an error as there are not 5 tokens in the bucket
	err = bucket.Take(5)
	// err.Error() => "Insufficient tokens."

	// put 5 tokens back into the bucket
	err = bucket.Put(5)
	// error == nil

	// wait for at least 10 tokens to be in the bucket (currently 5)
	done := bucket.Watch(10, time.Second * 5)
	// error == nil

	// put 5 tokens into the bucket
	err = bucket.Put(100)
	// error == nil

	// listen for bucket.Watch to return via the returned channel
	err = <- done

	// (err == nil)
	fmt.Println(err)
}