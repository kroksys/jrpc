package registry

// Service represents struct with its methods and is registered with a name
type Service struct {
	Name          string
	methods       map[string]*Method
	subscriptions map[string]*Method
}
