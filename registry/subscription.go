package registry

import (
	"github.com/kroksys/jrpc/conn"
	"github.com/kroksys/jrpc/spec"
)

// Subscription should be attached to any function that is ment to serve as a
// subscription. Just includine "func x(sub *Subscription) error" will mean
// that it will be used as subscription and should block the thread while its
// used.
type Subscription struct {
	methodName string
	Conn       *conn.Conn
}

// Creates new Subscription with its name and write channel.
// Returns nil if chanel is not provided.
func NewSubscription(methodName string, c ...*conn.Conn) *Subscription {
	if len(c) != 1 {
		return nil
	}
	return &Subscription{
		methodName: methodName,
		Conn:       c[0],
	}
}

// Sends json-rpc Notification to the open connection.
func (s *Subscription) Notify(data interface{}) bool {
	n := spec.NewNotification()
	n.Method = s.methodName
	n.Params = data
	if !s.Conn.Stopped {
		s.Conn.Write <- n
		return true
	}
	return false
}
