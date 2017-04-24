[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/b3ntly/bucket/master/LICENSE.txt) [![Build Status](https://travis-ci.org/b3ntly/bucket.svg?branch=master)](https://travis-ci.org/b3ntly/bucket)
[![Coverage Status](https://coveralls.io/repos/github/b3ntly/bucket/badge.svg?branch=master)](https://coveralls.io/github/b3ntly/bucket?branch=master?q=1) [![GoDoc](https://godoc.org/github.com/b3ntly/bucket?status.svg)](https://godoc.org/github.com/b3ntly/bucket)

## Bucket primitives with support for in-memory or Redis based storage

![An image of a bucket should go here](https://raw.githubusercontent.com/b3ntly/bucket/master/images/bucket.jpg)

The bucket is a simple and powerful tool. Don't doubt the bucket. With a bucket
you can:

* Implement token bucket algorithms
* Work with distributed systems
* Build sand castles


## Install

```golang
go get github.com/b3ntly/bucket
```

## Basic Usage

```golang
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

	// wait for up to 5 seconds for 10 tokens to be available
	done := b.Watch(10, time.Second * 5).Done()
	time.Sleep(time.Second * 6)
	// error == nil

	// put 100 tokens into the bucket
	err = b.Put(100)
	// error == nil

	// listen for bucket.Watch to return via the returned channel
	err = <- done
	// error == nil

	// (err == nil)
	fmt.Println(err)
}
```

## Redis

```golang
package main

import (
	"github.com/b3ntly/bucket"
	"github.com/b3ntly/bucket/storage"
	"github.com/go-redis/redis"
	"fmt"
)

func main(){
	storageOptions := &redis.Options{
		Addr:     "127.0.0.1:6379",
		PoolSize: 30,
	}

	store, err := storage.NewStorage("redis", storageOptions)

	if err != nil {
		// handle error
	}

	b, err := bucket.NewBucket("some_bucket", 10, store)

	if err != nil {
		// handle error
	}

	fmt.Println(b.Name)
}
```

## Multi-bucket

```golang
package main

import (
	tb "github.com/b3ntly/distributed-token-bucket"
	"github.com/go-redis/redis"
	"fmt"
)

func main(){
	var err error

	storageOptions := &redis.Options{
		Addr:     "127.0.0.1:6379",
		PoolSize: 30,
	}

	storage, err := tb.NewStorage("redis", storageOptions)
	// error == nil

	// you can create multiple buckets with the same storage instance to share client connections
	bucketOne, err := tb.NewBucket("bucket_one", 50, storage)
	bucketTwo, err := tb.NewBucket("bucket_two", 50, storage)
	bucketThree, err := tb.NewBucket("bucket_three", 50, storage)

	err = bucketOne.Take(5)
	err = bucketTwo.Take(5)
	err = bucketThree.Take(5)

	fmt.Println(err)
}
```

## Watchables

```golang
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
```

## Notes

* Test coverage badge is stuck in some cache and is out of date, click the badge to see the actual current coverage


## Changelog for version 0.2

* Abstracted storage to its own interface, see storage.go and redis.go for examples
* Added in-memory storage option
* Replaced storageOptions parameter with IStorage, allowing a single storage option to be shared
  between multiple buckets. This should make it much more efficient to have a large number of buckets, i.e.
  per logged in user.
  
# Changelog for version 0.3
  
* Renamed the repository from distributed-token-bucket to bucket
* Moved storage interface and providers to the /storage subpackage
* Added unit testing to the /storage subpackage
* Added watchable.go and changed signatures of all async functions to return a watchable
* Fixed examples
* Added more documentation and comments  

## Benchmarks

```golang
go test -bench .
```

These benchmarks are fairly incomplete and should be taken with a shot of tequila.


Version 0.1


| Benchmark                | Operations | ns/op  |
|--------------------------|------------|--------|
| BenchmarkBucket_Create-8 | 10000      | 139613 |
| BenchmarkBucket_Take-8   | 30000      | 40868  |
| BenchmarkBucket_Put-8    | 50000      | 29234  |

Version 0.2

Memory

| Benchmark                | Operations | ns/op  |
|--------------------------|------------|--------|
| BenchmarkBucket_Create-8 | 10000      | 605 ns/op |
| BenchmarkBucket_Take-8   | 30000      | 100 ns/op  |
| BenchmarkBucket_Put-8    | 50000      | 105 ns/op  |

Redis

| Benchmark                | Operations | ns/op  |
|--------------------------|------------|--------|
| BenchmarkBucket_Create-8 | 10000      | 71259 ns/op |
| BenchmarkBucket_Take-8   | 30000      | 47357 ns/op  |
| BenchmarkBucket_Put-8    | 50000      | 28360 ns/op  |

Version 0.3

Memory

| Benchmark                | Operations | ns/op  |
|--------------------------|------------|--------|
| BenchmarkBucket_Create-8 | 10000      | 572 ns/op |
| BenchmarkBucket_Take-8   | 30000      | 105 ns/op  |
| BenchmarkBucket_Put-8    | 50000      | 105 ns/op  |

Redis

| Benchmark                | Operations | ns/op  |
|--------------------------|------------|--------|
| BenchmarkBucket_Create-8 | 10000      | 75309 ns/op |
| BenchmarkBucket_Take-8   | 30000      | 49815 ns/op  |
| BenchmarkBucket_Put-8    | 50000      | 30638 ns/op  |



