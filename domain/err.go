package domain

import "errors"

var (
	ErrNotImplemented          = errors.New("Not implemented")
	ErrInvalidName             = errors.New("Invalid name. Name must only contain: A-Z, a-z, 0-9, -_")
	ErrInsufficientPermissions = errors.New("Insufficient permissions")
	ErrPayloadSizeExceeded     = errors.New("Payload size exceeds maximum allowed limit")
	ErrEmptyTopic              = errors.New("Topic cannot be empty")
	ErrEmptyTopicList          = errors.New("At least one topic must be specified")
)
