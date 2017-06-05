package lager

import (
	"context"
	"time"
)

type Context interface {
	context.Context
	Logger
}

type ctx struct {
	context.Context
	Logger
}

func NewContext(parent context.Context, logger Logger) Context {
	return &ctx{
		parent, logger,
	}
}

func WithCancel(parent Context) (Context, context.CancelFunc) {
	child, cancel := context.WithCancel(parent)
	return &ctx{
		child, parent,
	}, cancel
}

func WithTimeout(parent Context, duration time.Duration) (Context, context.CancelFunc) {
	child, cancel := context.WithTimeout(parent, duration)
	return &ctx{
		child, parent,
	}, cancel
}
