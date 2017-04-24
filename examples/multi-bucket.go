package main

import (
	"github.com/b3ntly/bucket"
	"github.com/b3ntly/bucket/storage"
	"fmt"
)

func main(){
	var err error

	store, err := storage.NewStorage("memory", nil)
	// error == nil

	// you can create multiple buckets with the same storage instance to share client connections
	bucketOne, err := bucket.NewBucket("bucket_one", 50, store)
	bucketTwo, err := bucket.NewBucket("bucket_two", 50, store)
	bucketThree, err := bucket.NewBucket("bucket_three", 50, store)

	err = bucketOne.Take(5)
	err = bucketTwo.Take(5)
	err = bucketThree.Take(5)

	fmt.Println(err)
}