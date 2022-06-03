package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/kroksys/jrpc/registry"
)

type Example struct{}

// {"jsonrpc":"2.0","method":"example.Simple", "id": 1, "params": [1, 2]}
func (Example) Simple(x, y int) (int, error) {
	return x + y, nil
}

type simpleRequest struct {
	X int
	Y int
}

// {"jsonrpc":"2.0","method":"example.SimpleObject", "id": 1, "params": {"Y": 1, "x": 2}}
func (Example) SimpleObject(req simpleRequest) (int, error) {
	log.Println("hello", req.X, req.Y)
	return req.X + req.Y, nil
}

// {"jsonrpc":"2.0","method":"example.Simple", "id": 1}
func (Example) SimpleNoParams() (int, error) {
	return 5, nil
}

func (Example) SimpleError() (int, error) {
	return 0, errors.New("simple error")
}

// {"jsonrpc":"2.0","method":"example.Simple", "id": 1, "params": [1, 2]}
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

func (Example) SubscriptionWithParams(
	ctx context.Context,
	sub *registry.Subscription,
	req simpleRequest,
) error {
	for i := 0; i < 10; i++ {
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

func (Example) SubWithComplexStruct(
	ctx context.Context,
	sub *registry.Subscription,
	req ComplexStruct,
) error {
	for i := 0; i < 2; i++ {
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

// !DONT DO THIS! - Params can havel only one struct passed so this will return error.
// Included only as an example.
func (Example) MultipleObject(req simpleRequest, multi int) (int, error) {
	return (req.X + req.Y) * multi, nil
}
