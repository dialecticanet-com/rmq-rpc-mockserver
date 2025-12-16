package expectations

import "errors"

var (
	ErrEmptyExchange   = errors.New("exchange cannot be empty")
	ErrEmptyRoutingKey = errors.New("routing key cannot be empty")
)
