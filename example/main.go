package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kroksys/jrpc"
)

func main() {
	host := "localhost:3333"
	gin.SetMode(gin.ReleaseMode)
	jrpcServer := jrpc.NewServer()
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
