package registry

type Service struct {
	Name          string
	methods       map[string]*Method
	subscriptions map[string]*Subscription
}
