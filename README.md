[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Build Status](https://travis-ci.org/b3ntly/distributed-token-bucket.svg?branch=master)](https://travis-ci.org/b3ntly/distributed-token-bucket) [![Coverage Status](https://coveralls.io/repos/github/b3ntly/distributed-token-bucket/badge.svg?branch=master)](https://coveralls.io/github/b3ntly/distributed-token-bucket?branch=master)

## Distributed Token Bucket with Redis and Golang

[example](./examples/server.go)

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
