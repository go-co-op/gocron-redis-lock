# redislock

[![CI State](https://github.com/go-co-op/gocron-redis-lock/actions/workflows/go_test.yml/badge.svg?branch=main&event=push)](https://github.com/go-co-op/gocron-redis-lock/actions)
![Go Report Card](https://goreportcard.com/badge/github.com/go-co-op/gocron-redis-lock) [![Go Doc](https://godoc.org/github.com/go-co-op/gocron-redis-lock?status.svg)](https://pkg.go.dev/github.com/go-co-op/gocron-redis-lock)

## install

```
go get github.com/go-co-op/gocron-redis-lock
```

## usage

Here is an example usage that would be deployed in multiple instances

```go
package main

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/redis/go-redis/v9"

	redislock "github.com/go-co-op/gocron-redis-lock"
)

func main() {
	redisOptions := &redis.Options{
		Addr: "localhost:6379",
	}
	redisClient := redis.NewClient(redisOptions)
	locker, err := redislock.NewRedisLocker(redisClient,time.Second * 2, redislock.WithTries(1))
	if err != nil {
		// handle the error
	}

	s := gocron.NewScheduler(time.UTC)
	s.WithDistributedLocker(locker)

	_, err = s.Every("1s").Name("unique_name").Do(func() {
		// task to do
		fmt.Println("call 1s")
	})
	if err != nil {
		// handle the error
	}

	s.StartBlocking()
}
```

The redis UniversalClient can also be used

```go
package main

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/redis/go-redis/v9"

	redislock "github.com/go-co-op/gocron-redis-lock"
)

func main() {
	redisOptions := &redis.UniversalOptions{
		Addrs: []string{"localhost:6379"},
	}
	redisClient := redis.NewUniversalClient(redisOptions)
	locker, err := redislock.NewRedisLocker(redisClient,time.Second * 2, redislock.WithTries(1))
	if err != nil {
		// handle the error
	}

	s := gocron.NewScheduler(time.UTC)
	s.WithDistributedLocker(locker)

	_, err = s.Every("1s").Name("unique_name").Do(func() {
		// task to do
		fmt.Println("call 1s")
	})
	if err != nil {
		// handle the error
	}

	s.StartBlocking()
}
```
