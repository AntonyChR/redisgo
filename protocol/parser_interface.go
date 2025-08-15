package protocol

type Parser interface{
	Encode(data string) ([]byte, error)
	EncodeBulkString(data string, addEnd bool) string
	Decode(data []byte) ([]string, error)
}
