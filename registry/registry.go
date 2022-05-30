package registry

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/kroksys/jrpc/conn"
	"github.com/kroksys/jrpc/spec"
	"github.com/kroksys/pool"
)

var (
	subscriptionType = reflect.TypeOf(&Subscription{})
	errorType        = reflect.TypeOf((*error)(nil)).Elem()
	contextType      = reflect.TypeOf((*context.Context)(nil)).Elem()
)

// Register struct as a service for jrpc-ws to handle automatically.
// Keeps track of subscriptions.
type Registry struct {

	// Registered services
	services *pool.PoolStr[Service]

	// Holds active subscriptions.
	// Subscription[key] - key = conn.Conn.ID + subscription.methodName
	subscriptions *pool.PoolStr[*Subscription]
}

// Creates new Registry with initialised services map
func NewRegistry() *Registry {
	return &Registry{
		services:      pool.NewPoolStr[Service](),
		subscriptions: pool.NewPoolStr[*Subscription](),
	}
}

// Call a method based on json-rpc Request. If a request is notification
// a Notification struct will be initialised and write channel attached to it.
// Returns response and ShouldReply flag.
func (reg *Registry) Call(ctx context.Context, req spec.Request, c *conn.Conn) (spec.Response, bool) {
	result := spec.NewResponse(req.ID, nil)
	split := strings.Split(req.Method, ".")
	if len(split) != 2 && len(split) != 3 {
		result.Error = spec.NewError(spec.MethodNotFoundCode, "invalid method name")
		return result, true
	}
	serviceName, methodName := split[0], strings.ToLower(split[1])
	var fn *Method
	var sub *Subscription
	if methodName == "subscribe" || methodName == "unsubscribe" {
		if len(split) == 3 {
			fn = reg.FindSubscription(serviceName, strings.ToLower(split[2]))
		} else {
			fn = reg.FindSubscription(serviceName)
		}
		if fn == nil {
			result.Error = spec.NewError(spec.InternalErrorCode, "invalid subscription name")
			return result, true
		}
		var ok bool
		sub, ok = reg.subscriptions.GetOk(c.ID + fn.name)
		if methodName == "subscribe" {
			if ok {
				result.Error = spec.NewError(spec.InternalErrorCode, "already subscribed")
				return result, true
			}
			sub = NewSubscription(fn.name, req.ID, c)
			reg.subscriptions.Put(sub.ID(), sub)
			defer reg.subscriptions.Delete(sub.ID())
		} else {
			if !ok {
				result.Error = spec.NewError(spec.InternalErrorCode, "not subscribed")
				return result, true
			}
			sub.Close()
			reg.subscriptions.Delete(sub.ID())
			return result, false
		}
	} else { // Method
		fn = reg.FindMethod(serviceName, methodName)
	}
	if fn == nil {
		result.Error = spec.NewError(spec.MethodNotFoundCode,
			fmt.Sprintf("missing services %s method %s", serviceName, methodName))
		return result, true
	}
	args, err := fn.ParseArgs(req.Params)
	if err != nil {
		result.Error = spec.NewError(spec.InvalidParamsCode, err.Error())
		return result, true
	}
	callResponse, err := fn.Call(ctx, methodName, args, sub)
	if err != nil {
		result.Error = spec.NewError(spec.InternalErrorCode, err.Error())
		return result, true
	}
	result.Result = callResponse
	return result, true
}

// Notify a method based on json-rpc Request. If a request is notification
// a Notification struct will be initialised and write channel attached to it.
func (reg *Registry) Subscribe(ctx context.Context, req spec.Notification, c *conn.Conn) *spec.Error {
	split := strings.Split(req.Method, ".")
	if len(split) != 2 && len(split) != 3 {
		return spec.NewError(spec.MethodNotFoundCode, "invalid method name")
	}
	serviceName, methodName := split[0], strings.ToLower(split[1])
	if methodName != "subscribe" && methodName != "unsubscribe" {
		return spec.NewError(spec.MethodNotFoundCode, "notifications can be used only to subscribe or unsubscribe")
	}

	var fn *Method
	if len(split) == 3 {
		fn = reg.FindSubscription(serviceName, strings.ToLower(split[2]))
	} else {
		fn = reg.FindSubscription(serviceName)
	}
	if fn == nil {
		return spec.NewError(spec.MethodNotFoundCode,
			fmt.Sprintf("missing subscription %s", req.Method))
	}

	sub, ok := reg.subscriptions.GetOk(c.ID + fn.name)
	switch methodName {
	case "subscribe":
		if ok {
			return spec.NewError(spec.InternalErrorCode, "already subscribled")
		}
		sub = NewSubscription(fn.name, nil, c)
		reg.subscriptions.Put(sub.ID(), sub)
		defer reg.subscriptions.Delete(sub.ID())
	case "unsubscribe":
		if !ok {
			return spec.NewError(spec.InternalErrorCode, "not subscribled")
		}
		sub.Close()
		reg.subscriptions.Delete(sub.ID())
		return nil
	}

	args, err := fn.ParseArgs(req.Params)
	if err != nil {
		return spec.NewError(spec.InvalidParamsCode, err.Error())
	}

	_, err = fn.Call(ctx, methodName, args, sub)
	if err != nil {
		return spec.NewError(spec.InternalErrorCode, err.Error())
	}
	return nil
}

// Register struct methods in registry. This should be called when server is
// initialised.
func (reg *Registry) Register(name string, service interface{}) error {
	methods, subscriptions := reg.extractMethods(reflect.ValueOf(service))
	if len(methods)+len(subscriptions) == 0 {
		return fmt.Errorf("service %T doesn't have methods to expose", service)
	}

	if _, ok := reg.services.GetOk(name); !ok {
		reg.services.Put(name, Service{
			Name:          name,
			methods:       methods,
			subscriptions: subscriptions,
		})
	}
	return nil
}

// Finds method in registry
func (reg *Registry) FindMethod(service, name string) *Method {
	return reg.services.Get(service).methods[name]
}

// Finds subscription in registry. Subscription in this case is just a method
// that can be called.
func (reg *Registry) FindSubscription(service string, name ...string) *Method {
	s := reg.services.Get(service)
	if len(name) == 1 {
		return s.subscriptions[name[0]]
	}
	if len(s.subscriptions) != 1 {
		return nil
	}
	for k := range s.subscriptions {
		return s.subscriptions[k]
	}
	return nil
}

// Extract functions/methods and subscriptions out of struct based on input and
// output parameters.
func (reg *Registry) extractMethods(theStruct reflect.Value) (map[string]*Method, map[string]*Method) {
	methods := make(map[string]*Method)
	subscriptions := make(map[string]*Method)
	structType := theStruct.Type()
	for i := 0; i < structType.NumMethod(); i++ {
		m := structType.Method(i)
		if m.PkgPath != "" { // not exported
			continue
		}
		fntype := m.Func.Type()
		// Arguments
		args := []reflect.Type{}
		hasCtx := false
		subPos := -1
		for j := 1; j < fntype.NumIn(); j++ {
			if j == 1 && fntype.In(j) == contextType {
				hasCtx = true
				continue
			}
			if fntype.In(j) == subscriptionType {
				subPos = j + 1
			}
			args = append(args, fntype.In(j))
		}
		// Returns
		errPos := -1
		if fntype.NumOut() > 2 {
			continue
		}
		for j := 0; j < fntype.NumOut(); j++ {
			if reg.isErrorType(fntype.Out(j)) {
				errPos = j
			}
		}
		meth := &Method{
			name:     m.Name,
			receiver: theStruct,
			fn:       m.Func,
			args:     args,
			errPos:   errPos,
			hasCtx:   hasCtx,
			subPos:   subPos,
		}
		if subPos != -1 {
			subscriptions[strings.ToLower(m.Name)] = meth
		} else {
			methods[strings.ToLower(m.Name)] = meth
		}
	}
	return methods, subscriptions
}

// Checks if type is an error
func (*Registry) isErrorType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Implements(errorType)
}
