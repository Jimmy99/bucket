[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Build Status](https://travis-ci.org/b3ntly/distributed-token-bucket.svg?branch=master)](https://travis-ci.org/b3ntly/distributed-token-bucket) [![Coverage Status](https://coveralls.io/repos/github/b3ntly/distributed-token-bucket/badge.svg?branch=master)](https://coveralls.io/github/b3ntly/distributed-token-bucket?branch=master)

## Distributed Token Bucket with Redis and Golang

Token-buckets are useful algorithms for things like rate-limiting and network congestion
control. Normally token-buckets are implemented in-memory which is helpful for high performance
applications. But what happens if you need rate limiting across a distributed system? This
library attempts to solve this problem by utilizing Redis as the token broker powering the
token bucket. Redis is particularly good at this because of it's relatively high level of
performance and concurrency control via only operating on a single thread.

## Install

```golang
go get github.com/b3ntly/distributed-token-bucket
```

## Basic Usage

```golang
package main

import (
    tb "github.com/b3ntly/distributed-token-bucket
    "github.com/go-redis/redis"
)

func main(){
    storageOptions = &redis.Options{
        Addr:     "127.0.0.1:6379",
        Password: "", // no password set
        DB:       1,  // use default DB
    }
    
    // initialize a bucket with 5 tokens
    bucket, err = tb.NewBucket(key, 5, storageOptions)
    
    // take 5 tokens
    err := bucket.Take(5)
    
    // try to take 5 tokens, this will return an error as there are not 5 tokens in the bucket
    err = bucket.Take(5)
    // err.Error() => "Insufficient tokens."
    
    // put 5 tokens back into the bucket
    err = bucket.Put(5)
    
    // wait for at least 10 tokens to be in the bucket (currently 5)
    done = bucket.Watch(10)
    
    // put 5 tokens into the bucket
    err = bucket.Put(5)
    
    // listen for bucket.Watch to return via the returned channel
    err = <- done
    
    // (err == nil)
}
```

## Notes

* Test coverage badge is stuck in some cache and is out of date, click the badge to see the actual current coverage


[real-life example](./examples/server.go)

## Benchmarks

These benchmarks are fairly incomplete due to a couple of reasons:

* Currently each instance of a bucket gets it's own redis.Client instance which do not
pool or share connections between them. This is really inefficient when creating large 
numbers of buckets.

* Most operations in this library are singular redis operations meaning the benchmark themselves are almost 
entirely pinned to the performance of Redis which is dominated by whatever
network latencies present between the instance and whatever process is 
utilizing this library.




| Benchmark                | Operations | ns/op  |
|--------------------------|------------|--------|
| BenchmarkBucket_Create-8 | 10000      | 139613 |
| BenchmarkBucket_Take-8   | 30000      | 40868  |
| BenchmarkBucket_Put-8    | 50000      | 29234  |
