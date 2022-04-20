package main

import (
	"context"
	"errors"
	"time"

	"github.com/kroksys/jrpc/registry"
)

type Example struct{}

func (Example) Simple(x, y int) (int, error) {
	return x + y, nil
}

func (Example) SimpleNoParams() (int, error) {
	return 5, nil
}

func (Example) SimpleError() (int, error) {
	return 0, errors.New("simple error")
}

func (Example) SimpleWithContext(ctx context.Context, x, y int) (int, error) {
	return x + y, nil
}

func (Example) Subscription(sub *registry.Subscription) error {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		sub.Notify("Hello")
	}
	return nil
}

func (Example) SubscriptionWithContext(ctx context.Context, sub *registry.Subscription) error {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		sub.Notify("Hello")
	}
	return errors.New("expected subscription break")
}
