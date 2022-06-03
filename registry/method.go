package registry

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"
	"strings"
)

// Method represents function in struct to be called.
// Registering struct with registry reflects all methods as Method
type Method struct {
	name     string
	receiver reflect.Value
	fn       reflect.Value
	args     []reflect.Type
	errPos   int
	hasCtx   bool
	subPos   int
}

// Transforms params interface coming from json parsed object to
// reflect values. It is neccessary to Call a Method.
func (m *Method) ParseArgs(params interface{}) ([]reflect.Value, error) {
	argCount := len(m.args)
	if argCount <= 0 {
		return []reflect.Value{}, nil
	}
	result := make([]reflect.Value, 0, argCount)
	switch reflect.ValueOf(params).Kind() {
	case reflect.Slice:
		if argCount != len(params.([]interface{})) {
			return nil, fmt.Errorf("arguments count does not match, expected %d arguments", len(m.args))
		}
		for i, param := range params.([]interface{}) {
			result = append(result, reflect.ValueOf(param).Convert(m.args[i]))
		}
	case reflect.Map:
		if argCount != 1 {
			return nil, fmt.Errorf("arguments count does not match, expected %d arguments", len(m.args))
		}
		argObject := reflect.New(m.args[0])
		argInterf := argObject.Interface()
		for k, v := range params.(map[string]interface{}) {
			err := SetInterfaceField(argInterf, k, v)
			if err != nil {
				return nil, fmt.Errorf("error decoding param object: %s", err.Error())
			}
		}
		result = append(result, argObject.Elem())
	}
	return result, nil
}

// Sets value (by name) of struct object created with reflect.New()
func SetInterfaceField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	name = strings.ToLower(name)
	structFieldValue := structValue.FieldByNameFunc(func(n string) bool {
		return strings.ToLower(n) == name
	})
	if !structFieldValue.IsValid() {
		return fmt.Errorf("no such field: %s in obj", name)
	}
	if !structFieldValue.CanSet() {
		return fmt.Errorf("cannot set %s field value", name)
	}
	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value).Convert(structFieldValue.Type())
	if structFieldType != val.Type() {
		return errors.New("provided value type didn't match obj field type")
	}
	structFieldValue.Set(val)
	return nil
}

// Executes function with given parameters. If a method is subscription it passes Subscription
// that holds write channel using Subscription.Notify().
func (m *Method) Call(ctx context.Context, method string, args []reflect.Value, sub *Subscription) (res interface{}, errRes error) {
	callArgs := []reflect.Value{m.receiver}
	if m.hasCtx {
		callArgs = append(callArgs, reflect.ValueOf(ctx))
	}
	if m.subPos != -1 {
		if sub == nil {
			return nil, errors.New("missing subscription implementation")
		}
		callArgs = append(callArgs, reflect.ValueOf(sub))
	}
	callArgs = append(callArgs, args...)

	// Catch panic
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Fprintln(ioutil.Discard, buf)
			// fmt.Println(string(buf))
			errRes = fmt.Errorf("%s:: %v", m.name, err)
		}
	}()

	// Run the callback.
	outputs := m.fn.Call(callArgs)
	if len(outputs) == 0 {
		return nil, nil
	}

	// Get error if exists
	if m.errPos >= 0 && !outputs[m.errPos].IsNil() {
		err := outputs[m.errPos].Interface().(error)
		return reflect.Value{}, err
	}

	return outputs[0].Interface(), nil
}
