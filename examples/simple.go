package main

import (
	"github.com/b3ntly/bucket"
	"time"
	"fmt"
)

func main(){
	// bucket will use in-memory storage as default
	b, err := bucket.New(&bucket.Options{
		Name: "my_bucket",
		Capacity: 10,
	})
	// err == nil

	// take 5 tokens
	err = b.Take(5)
	// err == nil

	// try to take 5 tokens, this will return an error as there are not 5 tokens in the bucket
	err = b.Take(5)
	// err.Error() => "Insufficient tokens."

	// put 5 tokens back into the bucket
	err = b.Put(5)
	// err == nil

	// watch for 10 tokens to be available, timing out after 5 seconds
	done := b.Watch(10, time.Second * 5).Done()
	// err == nil

	// put 5 tokens into the bucket
	err = b.Put(100)
	// error == nil

	// listen for bucket.Watch to return via the returned channel, it will return nil if 10 tokens could be acquired
	// else it will return an error from timing out, a manual cancelation (see ./watchable.go) or an actual error
	err = <- done
	// error == nil

	// will fill the bucket at the given rate when the interval channel is sent to
	signal := make(chan time.Time)
	watchable := b.DynamicFill(100, signal)
	signal <- time.Now()

	// stop the bucket from filling any longer
	watchable.Close(nil)

	// take all the tokens out of the bucket
	tokens, err := b.TakeAll()

	// (err == nil)
	fmt.Println(err)
	fmt.Println(tokens)
}