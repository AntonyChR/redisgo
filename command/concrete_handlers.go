package command

import (
	"context"
	"errors"
	"fmt"
	"net"
	protocol "redisgo/protocol"
	storage "redisgo/storage"
	"strconv"
	"strings"
	"time"
)

// PING
type PingHandler struct{}

func (p *PingHandler) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	_, err := conn.Write([]byte("+PONG\r\n"))
	return err
}

// ECHO
type EchoHandler struct {
	Parser protocol.Parser
}

func (e *EchoHandler) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	if len(args) == 0 {
		return errors.New("ECHO command requires an argument")
	}
	_, err := conn.Write([]byte(e.Parser.EncodeBulkString(args[0], true)))
	return err
}

// GET
type GetHandler struct {
	Storage *storage.Storage
	Parser  protocol.Parser
}

func (g *GetHandler) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	value, ok := g.Storage.Get(args[0])
	if !ok {
		_, err := conn.Write(nilResponse()) // Return null bulk string for non-existing key
		return err
	}
	encondedResp := g.Parser.EncodeBulkString(value, true)
	_, err := conn.Write([]byte(encondedResp))
	return err
}

// SET
type SetHandler struct {
	Storage     *storage.Storage
	ReplicaChan chan []byte
}

func (s *SetHandler) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	if len(args[0]) == 0 {
		_, err := conn.Write([]byte("-ERR invalid key value\r\n"))
		return err
	}
	key := args[0]
	value := args[1]
	s.Storage.Set(key, value)
	if len(args) == 4 {
		var expTime time.Duration
		if strings.ToLower(args[2]) == protocol.PX {
			t, err := strconv.Atoi(args[3])
			if err != nil {
				return errors.New("invalid expire time")
			}
			expTime = time.Duration(t) * time.Millisecond
		}

		if strings.ToLower(args[2]) == protocol.EX {
			t, err := strconv.Atoi(args[3])
			if err != nil {
				return errors.New("invalid expire time")
			}
			expTime = time.Duration(t) * time.Second
		}

		go func() {
			time.Sleep(expTime)
			s.Storage.DeleteValue(key)
		}()
	}

	_, err := conn.Write(okResponse())
	return err
}

// LRANGE
type LRange struct {
	Storage *storage.Storage
	Parser  protocol.Parser
}

func (l *LRange) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	start, _ := strconv.Atoi(args[1])
	stop, _ := strconv.Atoi(args[2])
	values := l.Storage.GetSliceFromList(args[0], start, stop)
	encoded := l.Parser.EncodeAsArray(values)
	_, err := conn.Write([]byte(encoded))
	return err
}

// LPUSH
type LPush struct {
	Storage *storage.Storage
}

func (l *LPush) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	if len(args[0]) == 0 {
		conn.Write([]byte("-ERR invalid key value\r\n"))
		return fmt.Errorf("empty key value")
	}

	key := args[0]
	values := args[1:]

	// invert values
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
	
	n := l.Storage.PrependValuesToList(key, values...)
	if len(values) > 0{
		l.Storage.NotifyWaiter(key, values[0])
	}
	resp := fmt.Sprintf(":%d\r\n", n)
	_, err := conn.Write([]byte(resp))
	return err
}

// RPUSH
type RPush struct {
	Storage *storage.Storage
}

func (s *RPush) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	if len(args[0]) == 0 {
		conn.Write([]byte("-ERR invalid key value\r\n"))
		return fmt.Errorf("empty key value")
	}

	key := args[0]
	values := args[1:]
	n := s.Storage.AppendValuesToList(key, values...)

	if n > 0  && n == len(values){
		s.Storage.NotifyWaiter(key, values[0])
	}
	resp := fmt.Sprintf(":%d\r\n", n)
	_, err := conn.Write([]byte(resp))
	return err
}

// LLEN

type LLEN struct {
	Storage *storage.Storage
}

func (l *LLEN) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	n := l.Storage.GetListLenght(args[0])
	resp := fmt.Sprintf(":%d\r\n", n)
	_, err := conn.Write([]byte(resp))
	return err

}

// LPOP
type LPOP struct {
	Storage *storage.Storage
	Parser  protocol.Parser
}

func (l *LPOP) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	resp := []byte{}
	if len(args) == 2 {
		n,_ := strconv.Atoi(args[1])
		values := l.Storage.RemoveFirstElementsFromTheList(args[0], n-1)
		resp = []byte(l.Parser.EncodeAsArray(values))
	} else {
		value := l.Storage.RemoveElementFromListByIndex(args[0], 0)
		resp = []byte(l.Parser.EncodeBulkString(value, true))
	}
	_, err := conn.Write([]byte(resp))
	return err
}

// BLPOP

type BLPOP struct {
	Storage *storage.Storage
	Parser protocol.Parser
}

func (b *BLPOP) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	key := args[0]
	value := b.Storage.RemoveElementFromListByIndex(key,0)
	if value != ""{
		t := []string{key, value}
		resp := []byte(b.Parser.EncodeAsArray(t))
		_, err := conn.Write([]byte(resp))
		return err
	}

	timeout := time.Hour

	if len(args) == 2{
		secs,_ := strconv.ParseFloat(args[1], 32)		
		if secs > 0 {
			milisecs := int(secs * 1000)
			timeout = time.Duration(milisecs) * time.Millisecond 
		}
	}

	waitChan := make(chan string, 1)
	b.Storage.RegisterWaiter(key, waitChan)
	defer func() {
		b.Storage.UnregisterWaiter(key, waitChan)
	}()

	select{
	case <- waitChan:
		val := b.Storage.RemoveElementFromListByIndex(key, 0) 
		if val != ""{
			t := []string{key, val}
			resp := []byte(b.Parser.EncodeAsArray(t))
			_, err := conn.Write([]byte(resp))
			return err
		}
		return nil
	case <- time.After(timeout):
		_, err := conn.Write(nilResponse())
		return err
	}
}

// TYPE

type Type struct {
	Storage *storage.Storage
	Parser  protocol.Parser
}


func (t *Type) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	valueType := t.Storage.CheckType(args[0])
	resp := []byte(t.Parser.EncodeAsSimpleString(valueType, true))
	_, err := conn.Write(resp)
	return err
}

type PSync struct {
}

func okResponse() []byte {
	return []byte("+OK\r\n")
}

func nilResponse() []byte {
	return []byte("$-1\r\n")
}
