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
		if !sub.IsRunning() {
			break
		}
		if err := sub.Notify("Hello"); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}

func (Example) SubscriptionWithContext(ctx context.Context, sub *registry.Subscription) error {
	for i := 0; i < 10; i++ {
		// <-sub.Exit and <-sub.Conn.Exit == !sub.IsRunning()
		select {
		case <-sub.Exit:
			return nil
		case <-sub.Conn.Exit:
			return nil
		default:
			if err := sub.Notify("Hello"); err != nil {
				return err
			}
			time.Sleep(time.Second)
		}
	}
	return errors.New("expected subscription break")
}
