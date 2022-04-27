package conn

import (
	"errors"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

// Websocket connection wrapper to handle JsonRpc communication
type Conn struct {
	ID        string
	c         net.Conn
	In        chan []byte
	Exit      chan interface{}
	closeOnce sync.Once
}

func NewConn(c net.Conn) *Conn {
	conn := Conn{
		ID:   uuid.NewString(),
		c:    c,
		In:   make(chan []byte),
		Exit: make(chan interface{}),
	}
	conn.GoRead()
	return &conn
}

// Sends ping message to the connection
func (c *Conn) Ping() {
	if !c.isRunning() {
		return
	}
	err := wsutil.WriteServerMessage(c.c, ws.OpPing, ws.CompiledPing)
	if err != nil {
		c.Close()
		return
	}
}

// Closes connection and its channels
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		wsutil.WriteServerMessage(c.c, ws.OpClose, nil)
		close(c.Exit)
		close(c.In)
	})
}

// gorutine for reading messages from connection
func (c *Conn) GoRead() {
	go func() {
		for {
			if !c.isRunning() {
				return
			}
			msg, _, err := wsutil.ReadClientData(c.c)
			if err != nil {
				c.Close()
				break
			}
			c.In <- msg
		}
	}()
}

// writes data to the connection
func (c *Conn) Send(msg []byte) error {
	if !c.isRunning() {
		return errors.New("cant send data to connection: connection is not running anymore")
	}
	err := wsutil.WriteServerMessage(c.c, ws.OpText, msg)
	if err != nil {
		c.Close()
	}
	return err
}

// Checks if connection is still running by reading from conn.Exit chanel.
// When connection is closed Exit chanel will be closed and will return false.
func (c *Conn) isRunning() bool {
	select {
	case _, running := <-c.Exit:
		return running
	default:
	}
	return true
}
