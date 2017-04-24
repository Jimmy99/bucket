package main

import (
	"github.com/b3ntly/bucket"
	"github.com/b3ntly/bucket/storage"
	"time"
	"fmt"
)

func main(){
	// use in-memory storage
	store, err := storage.NewStorage("memory", nil)
	// error == nil

	// initialize a bucket with 5 tokens
	b, err := bucket.NewBucket("simple_bucket", 5, store)

	// take 5 tokens
	err = b.Take(5)
	// error == nil

	// try to take 5 tokens, this will return an error as there are not 5 tokens in the bucket
	err = b.Take(5)
	// err.Error() => "Insufficient tokens."

	// put 5 tokens back into the bucket
	err = b.Put(5)
	// error == nil

	// watch for 10 tokens to be available, timing out after 5 seconds
	done := b.Watch(10, time.Second * 5).Done()
	// error == nil

	// put 5 tokens into the bucket
	err = b.Put(100)
	// error == nil

	// listen for bucket.Watch to return via the returned channel, it will return nil if 10 tokens could be acquired
	// else it will return an error from timing out, a manual cancelation (see ./watchable.go) or an actual error
	err = <- done
	// error == nil

	// (err == nil)
	fmt.Println(err)
}