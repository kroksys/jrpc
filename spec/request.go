package spec

const (
	JsonRpcVersion = "2.0"
)

type Request struct {

	// JSON-RPC protocol. MUST be exactly "2.0"
	Jsonrpc string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked
	Method string `json:"method"`

	// A Structured value that holds the parameter values to be
	// used during the invocation of the method.
	//
	// This member MAY be omitted.
	//
	// Array or Object from javascript JSON
	Params interface{} `json:"params"`

	// An identifier established by the Client that MUST contain
	// a String, Number, or NULL value
	// (this implementation ignores NULL)
	//
	// If it is not included it is assumed to be a notification.
	//
	// The Server MUST reply with the same value in the Response object if included.
	ID interface{} `json:"id,omitempty"`
}

// Checks if request is a notification
func (r *Request) IsNotification() bool {
	return r.ID == nil
}

// Returns new Request object with added JsonRpc version
func NewRequest() Request {
	return Request{
		Jsonrpc: JsonRpcVersion,
	}
}

// Decodes byte slice to Request object and returns pointer to it.
// If the data was not compatible with an object this func will return nil
func ParseRequest(data []byte) *Request {
	return fromBytes[Request](data, TypeRequest)
}
