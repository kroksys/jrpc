package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kroksys/jrpc/registry"
)

func NewGinWebsocketServer(host string) {
	gin.SetMode(gin.ReleaseMode)
	jrpcServer := NewServer()
	if err := jrpcServer.Register("example", Example{}); err != nil {
		log.Panicln(err)
	}
	r := gin.Default()
	r.GET("/ws", gin.WrapF(jrpcServer.WebsocketHandler))
	log.Printf("JSON RPC 2.0 server started. Address: %s/ws\n", host)
	err := r.Run(host)
	if err != nil {
		log.Println("jrpc server stopped")
	}
}

type Example struct{}

func (Example) Simple(x, y int) (int, error) {
	return x + y, nil
}

func (Example) SimpleWithContext(ctx context.Context, x, y int) (int, error) {
	return x + y, nil
}

func (Example) Subscription(ctx context.Context) (registry.Subscription, error) {
	return registry.Subscription{}, nil
}
