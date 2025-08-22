package protocol

type Parser interface{
	Encode(data []string) (string, error)
	EncodeBulkString(data string, addEnd bool) string
	EncodeAsSimpleString(data string, addEnd bool) string
	EncodeError(msg string) []byte
	EncodeAsArray(data []string) string
	NullBulkString() []byte 
	Ok() []byte
	Decode(data []byte) ([]string, error)
}
