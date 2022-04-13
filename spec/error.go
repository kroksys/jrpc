package spec

type Error struct {

	// A Number that indicates the error type that occurred.
	// This MUST be an integer.
	Code ErrorCode `json:"code"`

	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message ErrorMsg `json:"message"`

	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	// The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.).
	Data interface{} `json:"data,omitempty"`
}

// Returns new Error object with provided ErrorCode and data.
// Sets Error message using ErrorCode
func NewError(code ErrorCode, data interface{}) *Error {
	return &Error{
		Code:    code,
		Message: ErrorMessage(code),
		Data:    data,
	}
}

// Decodes byte slice to Error object and returns pointer to it.
// If the data was not compatible with an object this func will return nil
func ParseError(data []byte) *Error {
	return fromBytes[Error](data, TypeError)
}

type ErrorCode int

const (
	ParseErrorCode     ErrorCode = -32700
	InvalidRequestCode ErrorCode = -32600
	MethodNotFoundCode ErrorCode = -32601
	InvalidParamsCode  ErrorCode = -32602
	InternalErrorCode  ErrorCode = -32603
)

type ErrorMsg string

const (
	ParseErrorMsg     ErrorMsg = "Parse error"
	InvalidRequestMsg ErrorMsg = "Invalid Request"
	MethodNotFoundMsg ErrorMsg = "Method not found"
	InvalidParamsMsg  ErrorMsg = "Invalid params"
	InternalErrorMsg  ErrorMsg = "Internal error"
	ServerErrorMsg    ErrorMsg = "Server error"
)

func ErrorMessage(code ErrorCode) ErrorMsg {
	switch code {
	case ParseErrorCode:
		return ParseErrorMsg
	case InvalidRequestCode:
		return InvalidRequestMsg
	case MethodNotFoundCode:
		return MethodNotFoundMsg
	case InvalidParamsCode:
		return InvalidParamsMsg
	case InternalErrorCode:
		return InternalErrorMsg
	default:
		return ServerErrorMsg
	}
}
