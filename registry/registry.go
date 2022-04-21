package registry

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/kroksys/jrpc/conn"
	"github.com/kroksys/jrpc/spec"
)

var (
	subscriptionType = reflect.TypeOf(&Subscription{})
	errorType        = reflect.TypeOf((*error)(nil)).Elem()
	contextType      = reflect.TypeOf((*context.Context)(nil)).Elem()
)

// Registry for registering struct methods and subscriptions
type Registry struct {
	services map[string]Service
	lock     sync.Mutex
}

// Creates new Registry with initialised services map
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]Service),
	}
}

// Call a method based on json-rpc Request. If a request is notification
// a Notification struct will be initialised and write channel attached to it.
func (reg *Registry) Call(ctx context.Context, req spec.Request, c ...*conn.Conn) spec.Response {
	result := spec.NewResponse(req.ID, nil)
	split := strings.Split(req.Method, "_")
	if len(split) != 2 && len(split) != 3 {
		result.Error = spec.NewError(spec.MethodNotFoundCode, "invalid method name")
		return result
	}
	serviceName, methodName := split[0], strings.ToLower(split[1])
	var fn *Method
	if methodName == "subscribe" {
		var subscriptionName = ""
		if len(split) == 3 {
			subscriptionName = split[2]
			fn = reg.FindSubscription(serviceName, subscriptionName)
		} else {
			fn = reg.FindSubscription(serviceName)
		}
	} else { // Method
		fn = reg.FindMethod(serviceName, methodName)
	}
	if fn == nil {
		result.Error = spec.NewError(spec.MethodNotFoundCode,
			fmt.Sprintf("missing services %s method %s", serviceName, methodName))
		return result
	}
	args, err := fn.ParseArgs(req.Params)
	if err != nil {
		result.Error = spec.NewError(spec.InvalidParamsCode, err.Error())
		return result
	}
	callResponse, err := fn.Call(ctx, methodName, args, NewSubscription(methodName, c...))
	if err != nil {
		result.Error = spec.NewError(spec.InternalErrorCode, err.Error())
		return result
	}
	result.Result = callResponse
	return result
}

// Register struct methods in registry. This should be called when server is
// initialised.
func (reg *Registry) Register(name string, service interface{}) error {
	methods, subscriptions := reg.extractMethods(reflect.ValueOf(service))
	if len(methods)+len(subscriptions) == 0 {
		return fmt.Errorf("service %T doesn't have methods to expose", service)
	}
	reg.lock.Lock()
	defer reg.lock.Unlock()
	if _, ok := reg.services[name]; !ok {
		reg.services[name] = Service{
			Name:          name,
			methods:       methods,
			subscriptions: subscriptions,
		}
	}
	return nil
}

// Finds method in registry
func (reg *Registry) FindMethod(service, name string) *Method {
	reg.lock.Lock()
	defer reg.lock.Unlock()
	return reg.services[service].methods[name]
}

// Finds subscription in registry. Subscription in this case is just a method
// that can be called.
func (reg *Registry) FindSubscription(service string, name ...string) *Method {
	reg.lock.Lock()
	defer reg.lock.Unlock()
	if len(name) == 1 {
		return reg.services[service].subscriptions[name[0]]
	}
	if len(reg.services[service].subscriptions) != 1 {
		return nil
	}
	for k := range reg.services[service].subscriptions {
		return reg.services[service].subscriptions[k]
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
		responseChanPos := -1
		for j := 1; j < fntype.NumIn(); j++ {
			if j == 1 && fntype.In(j) == contextType {
				hasCtx = true
				continue
			}
			if fntype.In(j) == subscriptionType {
				responseChanPos = j + 1
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
			receiver: theStruct,
			fn:       m.Func,
			args:     args,
			errPos:   errPos,
			hasCtx:   hasCtx,
			chanPos:  responseChanPos,
		}
		if responseChanPos != -1 {
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
