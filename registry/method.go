package registry

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

type Method struct {
	receiver reflect.Value
	fn       reflect.Value
	args     []reflect.Type
	errPos   int
	hasCtx   bool
}

func (m *Method) ParseArgs(params interface{}) ([]reflect.Value, error) {
	result := make([]reflect.Value, 0, len(m.args))
	switch reflect.ValueOf(params).Kind() {
	case reflect.Slice:
		if len(m.args) != len(params.([]interface{})) {
			return nil, fmt.Errorf("arguments count does not match, expected %d arguments", len(m.args))
		}
		for i, param := range params.([]interface{}) {
			result = append(result, reflect.ValueOf(param).Convert(m.args[i]))
		}
	}
	return result, nil
}

func (m *Method) Call(ctx context.Context, method string, args []reflect.Value) (res interface{}, errRes error) {
	callArgs := []reflect.Value{m.receiver}
	if m.hasCtx {
		callArgs = append(callArgs, reflect.ValueOf(ctx))
	}
	callArgs = append(callArgs, args...)

	// Catch panic
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			buf = nil
			errRes = errors.New("method handler crashed")
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
