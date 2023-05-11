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
	"time"

	"github.com/go-co-op/gocron"
	redislock "github.com/go-co-op/gocron-redis-lock"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisOptions := &redis.Options{
		Addr: "localhost:6379",
	}
	redisClient := redis.NewClient(redisOptions)
	locker, err := redislock.NewRedisLocker(redisClient)
	if err != nil {
		// handle the error
	}

	s := gocron.NewScheduler(time.UTC)
	s.WithDistributedLocker(locker)

	_, err = s.Every("1s").Name("unique_name").Do(func() {
		// task to do
	})
	if err != nil {
		// handle the error
	}

	s.StartBlocking()
}
```
