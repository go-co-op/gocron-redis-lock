package redislock

import (
	"context"
	"fmt"
	"time"

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
func NewRedisLocker(r redis.UniversalClient, autoExtendDuration time.Duration, options ...redsync.Option) (gocron.Locker, error) {
	if err := r.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", gocron.ErrFailedToConnectToRedis, err)
	}
	return newLocker(r, autoExtendDuration, options...), nil
}

// NewRedisLockerAlways provides an implementation of the Locker interface using
// redis for storage, even if the connection fails.
func NewRedisLockerAlways(r redis.UniversalClient, autoExtendDuration time.Duration, options ...redsync.Option) (gocron.Locker, error) {
	return newLocker(r, autoExtendDuration, options...), r.Ping(context.Background()).Err()
}

func newLocker(r redis.UniversalClient, autoExtendDuration time.Duration, options ...redsync.Option) gocron.Locker {
	pool := goredis.NewPool(r)
	rs := redsync.New(pool)
	return &redisLocker{rs: rs, autoExtendDuration: autoExtendDuration, options: options}
}

var _ gocron.Locker = (*redisLocker)(nil)

type redisLocker struct {
	rs                 *redsync.Redsync
	options            []redsync.Option
	autoExtendDuration time.Duration
}

func (r *redisLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	mu := r.rs.NewMutex(key, r.options...)
	err := mu.LockContext(ctx)
	if err != nil {
		return nil, gocron.ErrFailedToObtainLock
	}
	rl := &redisLock{
		mu:                 mu,
		autoExtendDuration: r.autoExtendDuration,
		done:               make(chan struct{}),
	}

	if r.autoExtendDuration > 0 {
		go func() { rl.doExtend() }()
	}
	return rl, nil
}

var _ gocron.Lock = (*redisLock)(nil)

type redisLock struct {
	mu                 *redsync.Mutex
	done               chan struct{}
	autoExtendDuration time.Duration
}

func (r *redisLock) Unlock(ctx context.Context) error {
	close(r.done)
	unlocked, err := r.mu.UnlockContext(ctx)
	if err != nil {
		return gocron.ErrFailedToReleaseLock
	}
	if !unlocked {
		return gocron.ErrFailedToReleaseLock
	}

	return nil
}

func (r *redisLock) doExtend() {
	ticker := time.NewTicker(r.autoExtendDuration)
	for {
		select {
		case <-r.done:
			return
		case <-ticker.C:
			_, err := r.mu.Extend()
			if err != nil {
				return
			}
		}
	}
}
