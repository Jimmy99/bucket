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

	// put 100 tokens into the bucket
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
```

## Redis

```golang
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
```

## Multi-bucket

```golang
package main

import (
	"github.com/b3ntly/bucket"
	"github.com/b3ntly/bucket/storage"
	"github.com/go-redis/redis"
	"fmt"
)

func main(){
	var err error

	// with custom redis options
	store := &storage.RedisStorage{
		Client: redis.NewClient(&redis.Options{
			Addr: ":6379",
			DB: 5,
			PoolSize: 30,
		}),
	}

	b, err := bucket.New(&bucket.Options{
		Capacity: 10,
		Name: "My redis bucket with custom config 1",
		Storage: store,
	})

	b2, err := bucket.New(&bucket.Options{
		Capacity: 10,
		Name: "My redis bucket with custom config 2",
		Storage: store,
	})

	b3, err := bucket.New(&bucket.Options{
		Capacity: 10,
		Name: "My redis bucket with custom config 3",
		Storage: store,
	})

	err = b.Take(5)
	err = b2.Take(5)
	err = b3.Take(5)

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

# Changelog for version 0.4

* Shortened "constructor" names
* Default options
* Better "constructor" signatures
* bucket.DynamicFill()
* bucket.TakeAll()

## Benchmarks

```golang
go test -bench .
```

These benchmarks are fairly incomplete and should be taken with a grain of salt.


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

Version 0.4

Memory

| Benchmark                | Operations | ns/op  |
|--------------------------|------------|--------|
| BenchmarkBucket_Create-8 | 10000      | 715 ns/op |
| BenchmarkBucket_Take-8   | 30000      | 132 ns/op  |
| BenchmarkBucket_Put-8    | 50000      | 142 ns/op  |

Redis

| Benchmark                | Operations | ns/op  |
|--------------------------|------------|--------|
| BenchmarkBucket_Create-8 | 10000      | 98582 ns/op |
| BenchmarkBucket_Take-8   | 30000      | 47716 ns/op  |
| BenchmarkBucket_Put-8    | 50000      | 31350 ns/op  |



