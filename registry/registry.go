package registry

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/kroksys/jrpc/spec"
)

var (
	subscriptionType = reflect.TypeOf(Subscription{})
	errorType        = reflect.TypeOf((*error)(nil)).Elem()
	contextType      = reflect.TypeOf((*context.Context)(nil)).Elem()
)

type Registry struct {
	services map[string]Service
	lock     sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]Service),
	}
}

func (reg *Registry) Call(ctx context.Context, req spec.Request) spec.Response {
	result := spec.NewResponse(req.ID, nil)
	split := strings.Split(req.Method, "_")
	if len(split) != 2 {
		result.Error = spec.NewError(spec.MethodNotFoundCode, "invalid method name")
		return result
	}
	serviceName, methodName := split[0], split[1]
	method := reg.FindMethod(serviceName, methodName)
	if method == nil {
		result.Error = spec.NewError(spec.MethodNotFoundCode,
			fmt.Sprintf("missing services %s method %s", serviceName, methodName))
		return result
	}
	args, err := method.ParseArgs(req.Params)
	if err != nil {
		result.Error = spec.NewError(spec.InvalidParamsCode, err.Error())
		return result
	}
	callResponse, err := method.Call(ctx, methodName, args)
	if err != nil {
		result.Error = spec.NewError(spec.InternalErrorCode, err.Error())
		return result
	}
	result.Result = callResponse
	return result
}

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

func (reg *Registry) FindMethod(service, name string) *Method {
	reg.lock.Lock()
	defer reg.lock.Unlock()
	return reg.services[service].methods[name]
}

func (reg *Registry) FindSubscription(service, name string) *Subscription {
	reg.lock.Lock()
	defer reg.lock.Unlock()
	return reg.services[service].subscriptions[name]
}

func (reg *Registry) extractMethods(theStruct reflect.Value) (map[string]*Method, map[string]*Subscription) {
	methods := make(map[string]*Method)
	subscriptions := make(map[string]*Subscription)
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
		for j := 1; j < fntype.NumIn(); j++ {
			if j == 1 && fntype.In(j) == contextType {
				hasCtx = true
				continue
			}
			args = append(args, fntype.In(j))
		}
		// Returns
		numOut := fntype.NumOut()
		errPos := -1
		isSubscription := false
		if numOut > 2 {
			continue
		}
		if numOut == 2 {
			if !reg.isErrorType(fntype.Out(1)) {
				continue
			}
			errPos = 1
			isSubscription = reg.isSubscriptionType(fntype.Out(0))
		}
		if numOut == 1 {
			if reg.isErrorType(fntype.Out(0)) {
				errPos = 0
			}
		}
		if isSubscription {
			subscriptions[strings.ToLower(m.Name)] = &Subscription{
				receiver: theStruct,
				fn:       m.Func,
				args:     args,
				errPos:   errPos,
				hasCtx:   hasCtx,
			}
		} else {
			methods[strings.ToLower(m.Name)] = &Method{
				receiver: theStruct,
				fn:       m.Func,
				args:     args,
				errPos:   errPos,
				hasCtx:   hasCtx,
			}
		}
	}
	return methods, subscriptions
}

func (*Registry) isErrorType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Implements(errorType)
}

func (*Registry) isSubscriptionType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t == subscriptionType
}
