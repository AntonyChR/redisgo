package protocol

type Parser interface{
	Encode(data []string) (string, error)
	EncodeBulkString(data string, addEnd bool) string
	EncodeAsSimpleString(data string, addEnd bool) string
	EncodeError(msg string) []byte
	ConcatenateArray(data []string) string 
	MapToArray(data map[string]string) string
	EncodeAsArray(data []string) string
	NullBulkString() []byte 
	Ok() []byte
	Decode(data []byte) ([]string, error)
}
