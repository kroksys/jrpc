package registry

import "github.com/kroksys/jrpc/spec"

type Subscription struct {
	methodName string
	write      chan<- spec.Notification
}

func NewSubscription(methodName string, write ...chan<- spec.Notification) *Subscription {
	if len(write) != 1 {
		return nil
	}
	return &Subscription{
		methodName: methodName,
		write:      write[0],
	}
}

func (s *Subscription) Notify(data interface{}) {
	n := spec.NewNotification()
	n.Method = s.methodName
	n.Params = data
	s.write <- n
}
