package registry

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/kroksys/jrpc/conn"
	"github.com/kroksys/jrpc/spec"
)

// Subscription should be attached to any function that is ment to serve as a
// subscription. Just includine "func x(sub *Subscription) error" will mean
// that it will be used as subscription and should block the thread while its
// used.
type Subscription struct {
	// Flag to turn on/off logs for server
	LogsOn bool

	// ID provided by client request
	MessageID interface{}

	// Pointer to connection used to send data
	Conn *conn.Conn

	// Exit chanel will be closed if an "unsubscribe" is called or on error
	Exit     chan interface{}
	exitOnce sync.Once

	// Executed method name for subscription
	methodName string
}

// Creates new Subscription with its name and write channel.
// Returns nil if chanel is not provided.
func NewSubscription(methodName string, id interface{}, c *conn.Conn, logsOn bool) *Subscription {
	if id == nil {
		id = methodName
	}
	return &Subscription{
		MessageID:  id,
		Conn:       c,
		Exit:       make(chan interface{}),
		methodName: methodName,
		LogsOn:     logsOn,
	}
}

func (s *Subscription) Close() {
	s.exitOnce.Do(func() {
		close(s.Exit)
	})
}

// Unique key used to keep track of subscriptions.
// Key = Conn.ID + sub.methodName
func (s *Subscription) ID() string {
	return s.Conn.ID + s.methodName
}

// When handling subscription from struct use this function as a safety check.
// If returns false connection is closed or sub unsubscribed and handler loop
// should break.
func (s *Subscription) IsRunning() bool {
	select {
	case _, running := <-s.Exit:
		return running
	case _, running := <-s.Conn.Exit:
		s.Close()
		return running
	default:
	}
	return true
}

// Sends json-rpc Notification to the open connection.
func (s *Subscription) Notify(data interface{}) error {
	n := spec.NewResponse(s.MessageID, data)

	responseData, err := json.Marshal(n)
	if err != nil {
		if s.LogsOn {
			log.Printf("Error:%s:json.Marshal error: %s\n", s.methodName, err.Error())
		}
		return err
	}

	if s.LogsOn {
		log.Printf("Response:%s Id:%v Result:%.*v\n", s.methodName, n.ID, 20, n.Result)
	}
	err = s.Conn.Send(responseData)
	if err != nil {
		s.Close()
	}
	return err
}
