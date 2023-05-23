package redislock

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

// alias options
var (
	WithExpiry         = redsync.WithExpiry
	WithDriftFactor    = redsync.WithDriftFactor
	WithGenValueFunc   = redsync.WithGenValueFunc
	WithRetryDelay     = redsync.WithRetryDelay
	WithRetryDelayFunc = redsync.WithRetryDelayFunc
	WithTimeoutFactor  = redsync.WithTimeoutFactor
	WithTries          = redsync.WithTries
	WithValue          = redsync.WithValue
)

// NewRedisLocker provides an implementation of the Locker interface using
// redis for storage.
func NewRedisLocker(r redis.UniversalClient, options ...redsync.Option) (gocron.Locker, error) {
	if err := r.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", gocron.ErrFailedToConnectToRedis, err)
	}

	pool := goredis.NewPool(r)
	rs := redsync.New(pool)
	return &redisLocker{rs: rs, options: options}, nil
}

var _ gocron.Locker = (*redisLocker)(nil)

type redisLocker struct {
	rs      *redsync.Redsync
	options []redsync.Option
}

func (r *redisLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	mu := r.rs.NewMutex(key, r.options...)
	err := mu.LockContext(ctx)
	if err != nil {
		return nil, gocron.ErrFailedToObtainLock
	}
	rl := &redisLock{
		mu: mu,
	}
	return rl, nil
}

var _ gocron.Lock = (*redisLock)(nil)

type redisLock struct {
	mu *redsync.Mutex
}

func (r *redisLock) Unlock(ctx context.Context) error {
	unlocked, err := r.mu.UnlockContext(ctx)
	if err != nil {
		return gocron.ErrFailedToReleaseLock
	}
	if !unlocked {
		return gocron.ErrFailedToReleaseLock
	}

	return nil
}
