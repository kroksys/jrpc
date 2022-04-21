package jrpc

import (
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
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
	Write    chan spec.Notification
}

func newConn(c net.Conn) *Conn {
	conn := Conn{
		C:     c,
		in:    make(chan []byte),
		out:   make(chan []byte),
		exit:  make(chan interface{}),
		Write: make(chan spec.Notification),
	}
	conn.goRead()
	return &conn
}

// Sends ping message to the connection
func (c *Conn) ping() {
	err := wsutil.WriteServerMessage(c.C, ws.OpPing, ws.CompiledPing)
	if err != nil {
		c.close()
		return
	}
}

// Closes connection and its channels
func (c *Conn) close() {
	c.exitOnce.Do(func() {
		wsutil.WriteServerMessage(c.C, ws.OpClose, nil)
		close(c.exit)
		close(c.in)
		close(c.out)
		close(c.Write)
	})
}

// gorutine for reading messages from connection
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

// writes data to the connection
func (c *Conn) write(msg []byte) {
	err := wsutil.WriteServerMessage(c.C, ws.OpText, msg)
	if err != nil {
		c.close()
		return
	}
}
