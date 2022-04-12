package spec

import (
	"bytes"
	"encoding/json"
)

// JrpcType represents all JsonRpc specification types
type JrpcType int

const (
	TypeNone JrpcType = iota
	TypeBatchRequest
	TypeBatchResponse
	TypeError
	TypeNotification
	TypeRequest
	TypeResponse
)

func JrpcTypeString(tp JrpcType) string {
	switch tp {
	case TypeBatchRequest:
		return "TypeBatchRequest"
	case TypeBatchResponse:
		return "TypeBatchResponse"
	case TypeError:
		return "TypeError"
	case TypeNotification:
		return "TypeNotification"
	case TypeRequest:
		return "TypeRequest"
	case TypeResponse:
		return "TypeResponse"
	}
	return "TypeNone"
}

// Converts byte slice to JsonRpc and returns its Object and Type.
// [Request, Response, Notification, Error, BatchRequest, BatchResponse, None]
func GetJrpcType(data []byte) (interface{}, JrpcType) {
	switch GetJsonType(data) {
	case TypeJsonArray:
		array := []map[string]interface{}{}
		if err := json.Unmarshal(data, &array); err != nil {
			return nil, TypeNone
		}
		if len(array) > 0 {
			switch getObjectType(array[0]) {
			case TypeRequest:
				return array, TypeBatchRequest
			case TypeResponse:
				return array, TypeBatchResponse
			}
		}
	case TypeJsonObject:
		fieldMap := map[string]interface{}{}
		if err := json.Unmarshal(data, &fieldMap); err != nil {
			return nil, TypeNone
		}
		return fieldMap, getObjectType(fieldMap)
	}
	return nil, TypeNone
}

// Checks if fieldMap is of type JsonRpc. This does not include Batch request and response.
// [Request, Response, Notification, Error, None]
func getObjectType(fieldMap map[string]interface{}) JrpcType {
	for field := range fieldMap {
		switch field {
		case "code", "message":
			return TypeError
		case "result", "error":
			return TypeResponse
		case "method", "params":
			_, hasId := fieldMap["id"]
			if hasId {
				return TypeRequest
			} else {
				return TypeNotification
			}
		}
	}
	return TypeNone
}

// JsonType represents json Array and Object
type JsonType int

const (
	TypeJsonInvalid JsonType = iota
	TypeJsonArray
	TypeJsonObject
)

func JsonTypeString(tp JsonType) string {
	switch tp {
	case TypeJsonArray:
		return "TypeJsonArray"
	case TypeJsonObject:
		return "TypeJsonObject"
	}
	return "TypeJsonInvalid"
}

// Checks if is json type [Array, Object, None]
func GetJsonType(data []byte) JsonType {

	// Get slice of data with optional leading whitespace removed.
	// See RFC 7159, Section 2 for the definition of JSON whitespace.
	x := bytes.TrimLeft(data, " \t\r\n")

	isArray := len(x) > 0 && x[0] == '['
	isObject := len(x) > 0 && x[0] == '{'

	switch {
	case isArray:
		return TypeJsonArray
	case isObject:
		return TypeJsonObject
	}
	return TypeJsonInvalid
}
