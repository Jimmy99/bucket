package main

import (
	"time"
	"errors"
	"github.com/b3ntly/bucket"
	"github.com/b3ntly/bucket/storage"
	"fmt"
)

func main(){
	// use in-memory storage
	store, _ := storage.NewStorage("memory", nil)
	// error == nil

	b, _ := bucket.NewBucket("simple_bucket", 5, store)
	// error == nil

	watchable := b.Watch(10, time.Second * 5)
	watchable.Cancel <- errors.New("I wasn't happy with this watcher :/")

	// capture the error as the watcher exits
	err := <- watchable.Done()

	fmt.Println(err)
}
