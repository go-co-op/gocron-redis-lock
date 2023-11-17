package redislock

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testcontainersredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestEnableDistributedLocking(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testcontainersredis.RunContainer(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	uri, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	resultChan := make(chan int, 10)
	f := func(schedulerInstance int) {
		resultChan <- schedulerInstance
	}

	redisClient := redis.NewClient(&redis.Options{Addr: strings.TrimPrefix(uri, "redis://")})
	l, err := NewRedisLocker(redisClient, WithTries(1))
	require.NoError(t, err)

	s1, err := gocron.NewScheduler(gocron.WithDistributedLocker(l))
	require.NoError(t, err)
	_, err = s1.NewJob(gocron.DurationJob(500*time.Millisecond), gocron.NewTask(f, 1))
	require.NoError(t, err)

	s2, err := gocron.NewScheduler(gocron.WithDistributedLocker(l))
	require.NoError(t, err)
	_, err = s1.NewJob(gocron.DurationJob(500*time.Millisecond), gocron.NewTask(f, 2))
	require.NoError(t, err)

	s3, err := gocron.NewScheduler(gocron.WithDistributedLocker(l))
	require.NoError(t, err)
	_, err = s1.NewJob(gocron.DurationJob(500*time.Millisecond), gocron.NewTask(f, 3))
	require.NoError(t, err)

	s1.Start()
	s2.Start()
	s3.Start()

	time.Sleep(1700 * time.Millisecond)

	require.NoError(t, s1.Shutdown())
	require.NoError(t, s2.Shutdown())
	require.NoError(t, s3.Shutdown())
	close(resultChan)

	var results []int
	for r := range resultChan {
		results = append(results, r)
	}
	assert.Len(t, results, 4)
}
