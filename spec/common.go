package spec

import (
	"github.com/mitchellh/mapstructure"
)

// Convert byte slice to T object. Returns nil if JrpcType = TypeNone.
func fromBytes[T any](data []byte, tp JrpcType) *T {
	obj, t := GetJrpcType(data)
	if t == tp {
		return fromMap[T](obj)
	}
	return nil
}

// Converts map to T object
func fromMap[T any](m interface{}) *T {
	var obj = new(T)
	mapstructure.Decode(m, obj)
	return obj
}

// This function is more as an example and its own implementation is preferred.
/*	// Example for own implementation:
	obj, tp := GetJrpcType(byteSlice) // would be wise to use switch{} on tp
	req := spec.Request{}
	mapstructure.Decode(obj, &req)
*/
//
// Decodes byte slice to one of JsonRpc types and returns data and JsonRpc type.
// for the working project.
/*	// Example using Array:
	dataArray := []byte(`[{"jsonrpc":"2.0","method":"calc_add","params":[7,3],"id":4418}]`)
	ob, tp := spec.GetObject(dataArray)
	if tp == spec.TypeBatchRequest {
		request := ob.(spec.Request)
		fmt.Println(request)
	}

	// Example using Object:
	dataObject := []byte(`{"jsonrpc":"2.0","method":"calc_add","params":[7,3],"id":4418}`)
	ob, tp := spec.GetObject(dataArray)
	if tp == spec.TypeRequest {
		request := ob.(spec.Request)
		fmt.Println(request)
	}
*/
func Parse(data []byte) (interface{}, JrpcType) {
	obj, tp := GetJrpcType(data)
	var res interface{}
	switch tp {
	case TypeBatchRequest:
		res = BatchRequest{}
	case TypeRequest:
		res = Request{}
	case TypeBatchResponse:
		res = BatchResponse{}
	case TypeResponse:
		res = Response{}
	case TypeError:
		res = Error{}
	case TypeNotification:
		res = Notification{}
	}
	mapstructure.Decode(obj, &res)
	return res, tp
}
