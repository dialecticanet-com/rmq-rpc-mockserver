package expectations

import (
	"errors"
	"time"
)

type ExpectationOption func(e *Expectation) error

var (
	ErrBadRequestTimes          = errors.New("expectation call times must be greater than or equal to 1")
	ErrTTLMustBeGreaterThanZero = errors.New("expectation ttl must be greater than 0")
)

func WithPriority(p int) ExpectationOption {
	return func(e *Expectation) error {
		e.Priority = p
		return nil
	}
}

func WithLimitedTimes(times uint32) ExpectationOption {
	return func(e *Expectation) error {
		if times < 1 {
			return ErrBadRequestTimes
		}

		e.Times = &Times{
			RemainingTimes: times,
			Unlimited:      false,
		}

		return nil
	}
}

func WithUnlimitedTimes() ExpectationOption {
	return func(e *Expectation) error {
		e.Times = &Times{
			RemainingTimes: 0,
			Unlimited:      true,
		}
		return nil
	}
}

func WithTimeToLive(ttl time.Duration) ExpectationOption {
	return func(e *Expectation) error {
		if ttl <= 0 {
			return ErrTTLMustBeGreaterThanZero
		}

		e.TimeToLive = &TimeToLive{
			TTL: ttl,
		}

		return nil
	}
}
