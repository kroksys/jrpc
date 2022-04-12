package spec

// If the batch rpc call itself fails to be recognized as an
// valid JSON or as an Array with at least one value,
// the response from the Server MUST be a single Response object.
//
// If there are no Response objects contained within the Response
// array as it is to be sent to the client, the server
// MUST NOT return an empty Array and should return nothing at all.
// I.E. Client send: "[]" => Server does not reply at all
type BatchRequest []Request

// Decodes byte slice to BatchRequest object and returns pointer to it.
// If the data was not compatible with an object this func will return nil
func ParseBatchRequest(data []byte) *BatchRequest {
	return fromBytes[BatchRequest](data, TypeBatchRequest)
}
