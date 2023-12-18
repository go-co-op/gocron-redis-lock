package redislock

import (
	"time"

	"github.com/go-redsync/redsync/v4"
)

type LockerOption func(*redisLocker)

func WithAutoExtendDuration(duration time.Duration) LockerOption {
	return func(locker *redisLocker) {
		locker.autoExtendDuration = duration
	}
}

// WithRedsyncOptions sets the redsync options.
func WithRedsyncOptions(options ...redsync.Option) LockerOption {
	return func(locker *redisLocker) {
		locker.options = options
	}
}
