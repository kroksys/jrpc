package conn

import (
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/kroksys/jrpc/spec"
)

// Websocket connection wrapper to handle JsonRpc communication
type Conn struct {
	ID       string
	C        net.Conn
	In       chan []byte
	Out      chan []byte
	Exit     chan interface{}
	exitOnce sync.Once
	Stopped  bool
	Write    chan spec.Notification
}

func NewConn(c net.Conn) *Conn {
	conn := Conn{
		ID:    uuid.NewString(),
		C:     c,
		In:    make(chan []byte),
		Out:   make(chan []byte),
		Exit:  make(chan interface{}),
		Write: make(chan spec.Notification),
	}
	conn.GoRead()
	return &conn
}

// Sends ping message to the connection
func (c *Conn) Ping() {
	if c.Stopped {
		return
	}
	err := wsutil.WriteServerMessage(c.C, ws.OpPing, ws.CompiledPing)
	if err != nil {
		c.Close()
		return
	}
}

// Closes connection and its channels
func (c *Conn) Close() {
	c.exitOnce.Do(func() {
		wsutil.WriteServerMessage(c.C, ws.OpClose, nil)
		c.Stopped = true
		close(c.Exit)
		close(c.In)
		close(c.Out)
		close(c.Write)
	})
}

// gorutine for reading messages from connection
func (c *Conn) GoRead() {
	go func() {
		for {
			if c.Stopped {
				return
			}
			msg, _, err := wsutil.ReadClientData(c.C)
			if err != nil {
				c.Close()
				break
			}
			c.In <- msg
		}
	}()
}

// writes data to the connection
func (c *Conn) Send(msg []byte) {
	if c.Stopped {
		return
	}
	err := wsutil.WriteServerMessage(c.C, ws.OpText, msg)
	if err != nil {
		c.Close()
		return
	}
}
