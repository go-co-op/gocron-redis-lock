# redislock

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
	locker, err := redislock.NewRedisLocker(redisClient, redislock.WithTries(1))
	if err != nil {
		panic(err)
	}

	s := gocron.NewScheduler(time.UTC)
	s.WithDistributedLocker(locker)

	_, err = s.Every("1s").Name("unique_name").Do(func() {
		// task to do
		fmt.Println("call 1s")
	})
	if err != nil {
		panic(err)
	}

	s.StartBlocking()
}
```
