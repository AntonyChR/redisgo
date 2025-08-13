package protocol

type Parser interface{
	Encode(string) ([]byte, error)
	Decode([]byte) ([]string, error)
}
