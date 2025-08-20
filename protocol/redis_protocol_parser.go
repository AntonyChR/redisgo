package protocol


import (
	"strconv"
)

type RedisProtocolParser struct{}

func (r *RedisProtocolParser) Encode(data []string) (string, error){
	if len(data) > 0 || data[0] == REPLCONF || data[0] == PING{
		return r.EncodeAsArray(data), nil
	} else {
		return "+" + data[0] + ENDL , nil
	}
}

// bulkString returns a Redis bulk string representation of the given data.
// If addEnd is true, a line break is added at the end of the string.
func (p *RedisProtocolParser) EncodeBulkString(data string, addEnd bool) string {
	content := "$" + strconv.Itoa(len(data)) + "\r\n" + data
	if addEnd {
		content += "\r\n"
	}
	return content
}

func (r *RedisProtocolParser) EncodeError(msg string) []byte {
	return []byte("-ERR " + msg + ENDL)
}

func (r *RedisProtocolParser) EncodeAsArray(data []string) string {
	if len(data) == 0 {
		return "*0\r\n"
	}
	content := "*" + strconv.Itoa(len(data)) + "\r\n" + "$" + strconv.Itoa(len(data[0])) + "\r\n" + data[0] + "\r\n"
	for _, arg := range data[1:]{
		content += "$" + strconv.Itoa(len(arg)) + "\r\n" + arg + "\r\n"
	}

	return content
}

func (r *RedisProtocolParser) NullBulkString() []byte {
 return []byte{36, 45, 49, 13, 10} // $-1\r\n
}

func (r *RedisProtocolParser) Ok() []byte {
 return []byte{43, 79, 75, 13, 10} // +OK\r\n
}

func (r *RedisProtocolParser) Decode(data []byte) ([]string, error){
	return parseData(data), nil
}


