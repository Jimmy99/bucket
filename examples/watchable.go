package main

import (
	"time"
	"errors"
	"github.com/b3ntly/bucket"
	"fmt"
)

func main(){
	b, _ := bucket.New(&bucket.Options{
		Name: "my_bucket",
		Capacity: 10,
	})
	// error == nil

	watchable := b.Watch(10, time.Second * 5)
	watchable.Cancel <- errors.New("I wasn't happy with this watcher :/")

	// capture the error as the watcher exits
	err := <- watchable.Done()

	fmt.Println(err)
}
