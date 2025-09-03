package rabbitmq

import "errors"

var (
	ErrConsumeFailed = errors.New("Failed to consume from producer")
)

