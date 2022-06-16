package subscribers

import (
	"github.com/kroksys/pool"
)

// Subscribers pool to handle subscribers for jrpc.subscription.
type subscribers[T any] struct {
	watchers *pool.PoolStr[*pool.Pool[chan T]]
}

// Create new subscriber pool to handle jrpc.subscriptions.
/*
Example where single user.ID can havel multiple connections receiving User updates.
It's possible instead of user.ID to use any other string (i.e. order.ID) to receive User updates.
	// Outside of subscription
	s := subscribers.New[User]()

	// In subscription
	ch := make(chan User)
	chId := s.Register(User.ID, ch)
	defer func() {
		close(ch)
		s.Delete(User.ID, chId)
	}()

	// for { subscription handler }
*/
func New[T any]() subscribers[T] {
	return subscribers[T]{
		watchers: pool.NewPoolStr[*pool.Pool[chan T]](),
	}
}

// Adds new chan to subscription pool and return unique connection ID.
func (s *subscribers[T]) Register(group string, c chan T) uint64 {
	p, ok := s.watchers.GetOk(group)
	if !ok {
		p = pool.NewPool[chan T]()
		s.watchers.Put(group, p)
	}
	return p.Put(c)
}

// Removes chan to subscription pool. Tipically used with defer func() {}().
func (s *subscribers[T]) Delete(group string, chanId uint64) {
	if p, ok := s.watchers.GetOk(group); ok {
		p.Delete(chanId)
		if len(p.Data()) == 0 {
			s.watchers.Delete(group)
		}
	}
}

// Notify group of subscribers
func (s *subscribers[T]) NotifyGroup(o T, group string) {
	if p, ok := s.watchers.GetOk(group); ok {
		p.Each(func(c chan T) {
			c <- o
		})
	}
}

// Notify all subscribers
func (s *subscribers[T]) NotifyAll(o T) {
	s.watchers.Lock()
	defer s.watchers.Unlock()
	for _, p := range s.watchers.Data() {
		for _, c := range p.Data() {
			c <- o
		}
	}
}

// Notify one specific subscriber [unique connection]
func (s *subscribers[T]) NotifyID(o T, id uint64) {
	s.watchers.Lock()
	defer s.watchers.Unlock()
	for _, p := range s.watchers.Data() {
		if c, ok := p.GetOk(id); ok {
			c <- o
			return
		}
	}
}
