package rabbitmq

import "errors"

var (
	ErrConsumeFailed = errors.New("failed to consume from producer")
)

