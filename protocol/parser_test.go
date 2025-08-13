package protocol

import (
	"bytes"
	"testing"
)

// test parse simple string
type TestCases struct {
	input    string
	expected string
}

func TestParseSimpleString(t *testing.T) {
	cases := []TestCases{
		{"+OK\r\n+echo", "OK"},
		{"+PONG\r\n+second", "PONG"},
		{"+Hello World\r\n", "Hello World"},
	}

	for _, c := range cases {
		splittedData := bytes.Split([]byte(c.input), END_LINE)
		result := parseSimpleString(splittedData)
		if result != c.expected {
			t.Errorf("Expected %s but got %s", c.expected, result)
		}
	}
}

// test parse bulk string

func TestParseBulkString(t *testing.T) {
	cases := []TestCases{
		{"$5\r\nhello\r\n+noise", "hello"},
		{"$4\r\nping\r\n$2\r\nhi", "ping"},
		{"$0\r\n\r\n", ""},
	}

	for _, c := range cases {
		splittedData := bytes.Split([]byte(c.input), END_LINE)
		result, _ := parseBulkString(splittedData)
		if result != c.expected {
			t.Errorf("Expected %s but got %s", c.expected, result)
		}
	}
}

// test parse array

func TestParseArray(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{"*0\r\n", []string{}},
		{"*1\r\n$4\r\nPING\r\n", []string{"PING"}},
		{"*2\r\n$5\r\nhello\r\n$5\r\nworld", []string{"hello", "world"}},
		{"*3\r\n$5\r\nhello\r\n$5\r\nworld\r\n$5\r\nagain", []string{"hello", "world", "again"}},
		{"*2\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n$3\r\nSET\r\n$5\r\nfruit\r\n$5\r\napple", []string{"SET", "key", "value", "set", "fruit", "apple"}},
		{"*3\r\n$3\r\nSET\r\n$3\r\nbaz\r\n$3\r\n789\r\n", []string{"SET", "baz", "789"}},
		{"*3\r\n$3\r\nSET\r\n$3\r\nbar\r\n$3\r\n456\r\n", []string{"SET", "bar", "456"}},
		{"*3\r\n$8\r\nreplconf\r\n$6\r\ngetack\r\n$1\r\n*\r\n", []string{"replconf", "getack", "*"}},
	}

	for _, c := range cases {
		splittedData := bytes.Split([]byte(c.input), END_LINE)
		result, _ := parseArray(splittedData)
		for i, v := range result {
			if v != c.expected[i] {
				t.Errorf("Expected %s but got %s", c.expected[i], v)
			}
		}
	}
}

func TestParseData(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{"+OK\r\n", []string{"OK"}},
		{"$5\r\nhello\r\n", []string{"hello"}},
		{"*1\r\n$4\r\nPING\r\n", []string{"PING"}},
		{"*2\r\n$5\r\nhello\r\n$5\r\nworld", []string{"hello", "world"}},
		{"*3\r\n$5\r\nhello\r\n$5\r\nworld\r\n$5\r\nagain", []string{"hello", "world", "again"}},
		{"*2\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n$3\r\nSET\r\n$5\r\nfruit\r\n$5\r\napple", []string{"SET", "key", "value", "SET", "fruit", "apple"}},
		{"*3\r\n$3\r\nSET\r\n$3\r\nbaz\r\n$3\r\n789\r\n", []string{"SET", "baz", "789"}},
		{"*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\n123\r\n*3\r\n$3\r\nSET\r\n$3\r\nbar\r\n$3\r\n456", []string{"SET", "foo", "123", "SET", "bar", "456"}},
		{"*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n", []string{"PSYNC", "?", "-1"}},
	}

	for _, c := range cases {
		result := parseData([]byte(c.input))

		// check if the length of the result is the same as the expected
		if len(result) != len(c.expected) {
			t.Errorf("Expected %v but got %v", len(c.expected), len(result))
		}

		for i, v := range result {
			if v != c.expected[i] {
				t.Errorf("Expected %s but got %s", c.expected[i], v)
			}
		}
	}
}

func TestEncodeData(t *testing.T) {
	cases := []struct{
		input []string
		expected string
	}{
		{ []string{"PING"}, "*1\r\n$4\r\nPING\r\n"},
		{ []string{"REPLCONF","listening-port","uno" }, "*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$3\r\nuno\r\n"},
		{ []string{"REPLCONF", "capa", "sync2"}, "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$5\r\nsync2\r\n" },
		{ []string{"PSYNC", "?","-1"}, "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n" },
	}

	parser := RedisProtocolParser{}

	for i, c := range cases {
		resp,err := parser.Econde(c.input)

		if err != nil{
			t.Errorf("Error encoding data: %s", err)
		}

		if resp != c.expected{
			t.Errorf("error in case [%d], expected: %s, got: %v", i, c.expected, resp)
		}
	}
}
