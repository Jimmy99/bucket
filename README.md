[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Build Status](https://travis-ci.org/b3ntly/distributed-token-bucket.svg?branch=master)](https://travis-ci.org/b3ntly/distributed-token-bucket) [![Coverage Status](https://coveralls.io/repos/github/b3ntly/distributed-token-bucket/badge.svg?branch=master)](https://coveralls.io/github/b3ntly/distributed-token-bucket?branch=master) 
[![GoDoc](https://godoc.org/github.com/b3ntly/distributed-token-bucket?status.svg)](https://godoc.org/github.com/b3ntly/distributed-token-bucket)


## Distributed Token Bucket with Redis and Golang

The token bucket algorithm is useful for things like rate-limiting and network congestion
control. 

Normally token buckets are implemented in-memory which is helpful for high performance
applications. But what happens if you need rate limiting across a distributed system? 

This library attempts to solve said problem by utilizing Redis as the token broker powering the
token bucket. Redis is particularly good at this because of its relatively high level of
performance and concurrency control via its single-threaded runtime.

## Install

```golang
go get github.com/b3ntly/distributed-token-bucket
```

## Basic Usage

```golang
package main

import (
	tb "github.com/b3ntly/distributed-token-bucket"
	"github.com/go-redis/redis"
	"time"
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

## Notes

* Test coverage badge is stuck in some cache and is out of date, click the badge to see the actual current coverage


## Changelog for version 0.2

* Abstracted storage to its own interface, see storage.go and redis.go for examples
* Added in-memory storage option
* Replaced storageOptions parameter with IStorage, allowing a single storage option to be shared
  between multiple buckets. This should make it much more efficient to have a large number of buckets, i.e.
  per logged in user.

## Examples

[simple](./examples/simple.go)

[multi-bucket](./examples/mult-bucket.go)

[http-server](./examples/server.go)

## Benchmarks

These benchmarks are fairly incomplete due to a couple of reasons:

* Most operations in this library are singular redis operations meaning the benchmark themselves are almost 
entirely pinned to the performance of Redis which is dominated by whatever
network latencies present between the instance and whatever process is 
utilizing this library.


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



