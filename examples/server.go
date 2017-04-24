package main

import (
	"fmt"
	tb "github.com/b3ntly/distributed-token-bucket"
	"github.com/go-redis/redis"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	bucket *tb.Bucket
	wg     = &sync.WaitGroup{}

	requests         int64 = 0
	requestsAccepted int64 = 0

	storageOptions = &redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	}

	netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	key = "some_key_007"
)

func main() {
	var err error

	storage, err := tb.NewStorage("redis", storageOptions)

	if err != nil {
		fmt.Println(err)
	}

	bucket, err = tb.NewBucket(key, 5, storage)

	if err != nil {
		fmt.Println(err)
	}

	go server()

	wg.Add(10)
	for i := 0; i < 10; i++ {
		client()
	}

	wg.Wait()

	cleanUp()
	fmt.Printf("%v requests handled, %v requests accepted.\n", atomic.LoadInt64(&requests), atomic.LoadInt64(&requestsAccepted))
}

func client() {
	resp, err := netClient.Get("http://127.0.0.1:8080/")
	defer resp.Body.Close()

	if err != nil {
		fmt.Println(err)
		// handle error
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	atomic.AddInt64(&requests, int64(1))

	err := bucket.Take(1)
	//fmt.Println(err)

	if err == nil {
		atomic.AddInt64(&requestsAccepted, int64(1))
	}

	io.WriteString(w, "ok")
	wg.Done()
}

func server() {
	http.HandleFunc("/", handler)
	go http.ListenAndServe(":8080", nil)
}

func cleanUp() {
	client := redis.NewClient(storageOptions)
	err := client.Del(key).Err()

	if err != nil {
		fmt.Println(err)
	}
}
