package redislock

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-co-op/gocron"
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

	s1 := gocron.NewScheduler(time.UTC)
	s1.WithDistributedLocker(l)
	_, err = s1.Every("500ms").Do(f, 1)
	require.NoError(t, err)

	s2 := gocron.NewScheduler(time.UTC)
	s2.WithDistributedLocker(l)
	_, err = s2.Every("500ms").Do(f, 2)
	require.NoError(t, err)

	s3 := gocron.NewScheduler(time.UTC)
	s3.WithDistributedLocker(l)
	_, err = s3.Every("500ms").Do(f, 3)
	require.NoError(t, err)

	s1.StartAsync()
	s2.StartAsync()
	s3.StartAsync()

	time.Sleep(1700 * time.Millisecond)

	s1.Stop()
	s2.Stop()
	s3.Stop()
	close(resultChan)

	var results []int
	for r := range resultChan {
		results = append(results, r)
	}
	assert.Len(t, results, 4)
}

func TestAutoExtend(t *testing.T) {
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

	redisClient := redis.NewClient(&redis.Options{Addr: strings.TrimPrefix(uri, "redis://")})
	// create lock not auto extend
	l1, err := NewRedisLockerWithOptions(redisClient, WithRedsyncOptions(WithTries(1)))
	_, err = l1.Lock(ctx, "test1")
	require.NoError(t, err)

	t.Logf("waiting 9 seconds for lock to expire")
	// wait for the lock to expire
	time.Sleep(9 * time.Second)

	_, err = l1.Lock(ctx, "test1")
	require.NoError(t, err)

	// create auto extend lock
	l2, err := NewRedisLockerWithOptions(redisClient, WithAutoExtendDuration(time.Second*2), WithRedsyncOptions(WithTries(1)))
	unlocker, err := l2.Lock(ctx, "test2")
	require.NoError(t, err)

	t.Log("waiting 9 seconds for lock to expire")
	// wait for the lock to expire
	time.Sleep(9 * time.Second)

	_, err = l2.Lock(ctx, "test2")
	require.Equal(t, gocron.ErrFailedToObtainLock, err)

	err = unlocker.Unlock(ctx)
	require.NoError(t, err)
}
