package redislock

import (
	"time"

	"github.com/go-redsync/redsync/v4"
)

// LockerOption defines a function type that can be implemented to add options to the Locker
type LockerOption func(*redisLocker)

// WithAutoExtendDuration sets the duration for auto extending the lock.
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
