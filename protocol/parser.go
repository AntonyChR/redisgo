package protocol

import (
	"bytes"
	"strconv"
)

// comands
const (
	GET        = "get"
	SET        = "set"
	INFO       = "info"
	PING       = "ping"
	ECHO       = "echo"
	PSYNC      = "psync"
	REPLCONF   = "replconf"
	FULLRESYNC = "fullresync"
	WAIT       = "wait"
)

const ENDL string ="\r\n"

// set params
const (
	EX = "ex" // seconds
	PX = "px" // milliseconds

	// temporal values, TODO: move to a config file
	BASE64_EMPTY_RDB_FILE = "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
)

type RedisProtocolParser struct{}

func (r *RedisProtocolParser) Econde(data []string) (string, error){
	if len(data) > 0 || data[0] == REPLCONF || data[0] == PING{
		return r.EncodeAsArray(data), nil
	} else {
		return "+" + data[0] + ENDL , nil
	}
}

func (r *RedisProtocolParser) EncodeAsArray(data []string) string {
	content := "*" + strconv.Itoa(len(data)) + "\r\n" + "$" + strconv.Itoa(len(data[0])) + "\r\n" + data[0] + "\r\n"
	for _, arg := range data[1:]{
		content += "$" + strconv.Itoa(len(arg)) + "\r\n" + arg + "\r\n"
	}

	return content
}

func (r *RedisProtocolParser) EncondeError(msg string) []byte {
	return []byte("-ERR " + msg + ENDL)
}

func (r *RedisProtocolParser) Ok() []byte {
 return []byte{43, 79, 75, 13, 10} // +OK\r\n
}


func (r *RedisProtocolParser) NullBulkString() []byte {
 return []byte{36, 45, 49, 13, 10} // $-1\r\n
}
func (r *RedisProtocolParser) Decode(data []byte) ([]string, error){
	return parseData(data), nil
}

const (
	SIMPLE_STRINGS   = byte('+')
	SIMPLE_ERRORS    = byte('-')
	INTEGERS         = byte(':')
	BULK_STRINGS     = byte('$')
	ARRAY            = byte('*')
	NULLS            = byte('_')
	BOOLEANS         = byte('#')
	DOUBLES          = byte(',')
	BIG_NUMBERS      = byte('(')
	BULK_ERRORS      = byte('!')
	VERBATIM_STRINGS = byte('=')
	MAPS             = byte('%')
	SETS             = byte('~')
	PUSHES           = byte('>')
)

var END_LINE = []byte("\r\n")

// this function tranform redis protocol data like "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\n123\r\n*3\r\n$3\r\nSET\r\n$3\r\nbar\r\n$3\r\n456"
// into a slice of strings like ["SET", "foo", "123", "SET", "bar", "456"]
func parseData(data []byte) []string {
	splittedData := bytes.Split(bytes.TrimSuffix(data, END_LINE), END_LINE)
	result := make([]string, 0)
	for i := 0; i < len(splittedData); i++ {

		if len(splittedData[i]) == 0 {
			continue
		}

		switch splittedData[i][0] {
		case SIMPLE_STRINGS:
			result = append(result, parseSimpleString(splittedData[i:]))
		case BULK_STRINGS:
			str, n := parseBulkString(splittedData[i:])
			result = append(result, str)
			i += n - 1 // - 1 because the loop will increment i by 1
		case ARRAY:
			arr, n := parseArray(splittedData[i:])
			result = append(result, arr...)
			i += n - 1
		}
	}
	return result
}

func parseSimpleString(data [][]byte) string {
	return string(data[0][1:])
}

// parseBulkString parses a bulk string from the given data and returns the string value and the number of elements consumed.
// The data parameter is a 2D byte slice where the first element represents the length of the bulk string and the second element represents the actual string data.
// The function returns the parsed string value and the number of elements consumed (2 in this case).
func parseBulkString(data [][]byte) (string, int) {
	bulkLength, _ := strconv.Atoi(string(data[0][1:]))
	return string(data[1][0:bulkLength]), 2
}

// parseArray parses an array from the given data and returns the array and the number of elements consumed.
// returns consumed because we need to know how many elements we have consumed in the data slice
func parseArray(data [][]byte) (arr []string, consumed int) {
	//[0] *3    -> data[0][1:] = arrLength = 3
	//[1] $5    -> since data[1][0] = $, it's a bulk string
	//[2] hello
	//[3] $5
	//[4] world
	//[5] $5
	//[6] again
	arrLength, _ := strconv.Atoi(string(data[0][1:]))
	arr = make([]string, 0)
	consumed = 1 // we have consumed the first element which is the array length
	count := 0
	for i := 1; i < len(data); i++ {

		// if count == arrLength, it means we have consumed all the elements in the array
		if count == arrLength {
			break
		}
		switch data[i][0] {
		case BULK_STRINGS:
			str, n := parseBulkString(data[i:])
			arr = append(arr, str)
			consumed += n
			i += n - 1 // - 1 because the loop will increment i by 1
			count++
		case SIMPLE_STRINGS:
			str := parseSimpleString(data[i:])
			arr = append(arr, str)
			consumed++
			count++
		}
	}

	return arr, consumed
}

// bulkString returns a Redis bulk string representation of the given data.
// If addEnd is true, a line break is added at the end of the string.
func (p *RedisProtocolParser) EcondeBulkString(data string, addEnd bool) string {
	content := "$" + strconv.Itoa(len(data)) + "\r\n" + data
	if addEnd {
		content += "\r\n"
	}
	return content
}
