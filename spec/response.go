package spec

type Response struct {
	// JSON-RPC protocol. MUST be exactly "2.0"
	Jsonrpc string `json:"jsonrpc"`

	// This member is REQUIRED on success.
	// This member MUST NOT exist if there was an error invoking the method.
	// The value of this member is determined by the method invoked on the Server.
	Result interface{} `json:"result,omitempty"`

	// This member is REQUIRED on error.
	// This member MUST NOT exist if there was no error triggered during invocation.
	// The value for this member MUST be an Object as defined in section 5.1.
	Error *Error `json:"error,omitempty"`

	// This member is REQUIRED.
	// It MUST be the same as the value of the id member in the Request Object.
	// If there was an error in detecting the id in the Request object
	// (e.g. Parse error/Invalid Request), it MUST be Null.
	ID interface{} `json:"id"`

	// Either the result member or error member MUST be included, but both members MUST NOT be included.
}

// Returns new Response object with provided ID and result object
func NewResponse(id interface{}, result interface{}) Response {
	return Response{
		Jsonrpc: JsonRpcVersion,
		ID:      id,
		Result:  result,
	}
}

// Returns new error Response object with provided ID and error object
func NewResponseError(id interface{}, err Error) Response {
	return Response{
		Jsonrpc: JsonRpcVersion,
		ID:      id,
		Error:   &err,
	}
}

// Decodes byte slice to Response object and returns pointer to it.
// If the data was not compatible with an object this func will return nil
func ParseResponse(data []byte) *Response {
	return fromBytes[Response](data, TypeResponse)
}
