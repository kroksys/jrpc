package registry

import "github.com/kroksys/jrpc/spec"

// Subscription should be attached to any function that is ment to serve as a
// subscription. Just includine "func x(sub *Subscription) error" will mean
// that it will be used as subscription and should block the thread while its
// used.
type Subscription struct {
	methodName string
	write      chan<- spec.Notification
}

// Creates new Subscription with its name and write channel.
// Returns nil if chanel is not provided.
func NewSubscription(methodName string, write ...chan<- spec.Notification) *Subscription {
	if len(write) != 1 {
		return nil
	}
	return &Subscription{
		methodName: methodName,
		write:      write[0],
	}
}

// Sends json-rpc Notification to the open connection.
func (s *Subscription) Notify(data interface{}) {
	n := spec.NewNotification()
	n.Method = s.methodName
	n.Params = data
	s.write <- n
}
