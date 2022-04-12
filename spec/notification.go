package spec

type Notification struct {
	// JSON-RPC protocol. MUST be exactly "2.0"
	Jsonrpc string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked
	Method string `json:"method"`

	// A Structured value (array or object) that holds the parameter values to be
	// used during the invocation of the method.
	//
	// This member MAY be omitted.
	//
	// Array or Object from javascript JSON
	Params string `json:"params"`
}

// Returns new notification object with JsonRpc version attached
func NewNotification() Notification {
	return Notification{
		Jsonrpc: JsonRpcVersion,
	}
}

// Decodes byte slice to Notification object and returns pointer to it.
// If the data was not compatible with an object this func will return nil
func ParseNotification(data []byte) *Notification {
	return fromBytes[Notification](data, TypeNotification)
}
