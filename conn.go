package jrpc

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/kroksys/jrpc/registry"
	"github.com/kroksys/jrpc/spec"
)

const (
	pingPeriod = 30 * time.Second
)

// Websocket connection wrapper to handle JsonRpc communication
type Conn struct {
	C        net.Conn
	in       chan []byte
	out      chan []byte
	exit     chan interface{}
	exitOnce sync.Once
	registry *registry.Registry
}

func newConn(c net.Conn, reg *registry.Registry) *Conn {
	conn := Conn{
		C:        c,
		in:       make(chan []byte),
		out:      make(chan []byte),
		exit:     make(chan interface{}),
		registry: reg,
	}
	conn.goRead()
	return &conn
}

func (c *Conn) defaultHandler() {
	pinger := time.NewTicker(pingPeriod)
	defer pinger.Stop()
	for {
		select {
		case msg := <-c.in:
			data, tp := spec.Parse(msg)
			switch tp {
			case spec.TypeRequest:
				request := data.(spec.Request)
				response := c.registry.Call(context.TODO(), request)
				responseData, err := json.Marshal(response)
				if err != nil {
					continue
				}
				go func() {
					c.out <- responseData
				}()
			}
		case msg := <-c.out:
			c.write(msg)
		case <-pinger.C:
			c.ping()
		case <-c.exit:
			return
		}
	}
}

func (c *Conn) ping() {
	err := wsutil.WriteServerMessage(c.C, ws.OpPing, ws.CompiledPing)
	if err != nil {
		c.close()
		return
	}
}

func (c *Conn) close() {
	c.exitOnce.Do(func() {
		wsutil.WriteServerMessage(c.C, ws.OpClose, nil)
		close(c.exit)
		close(c.in)
		close(c.out)
	})
}

func (c *Conn) goRead() {
	go func() {
		for {
			msg, _, err := wsutil.ReadClientData(c.C)
			if err != nil {
				c.close()
				break
			}
			c.in <- msg
		}
	}()
}

func (c *Conn) write(msg []byte) {
	err := wsutil.WriteServerMessage(c.C, ws.OpText, msg)
	if err != nil {
		c.close()
		return
	}
}
