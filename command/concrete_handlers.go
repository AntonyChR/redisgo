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
			s.Storage.Delete(key)
		}()
	}

	_, err := conn.Write(okResponse())
	return err
}

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
	n := s.Storage.SetValueToList(key, values...)
	resp := fmt.Sprintf(":%d\r\n", n)
	_, err := conn.Write([]byte(resp))
	return err
}

type LRange struct {
	Storage *storage.Storage
	Parser protocol.Parser
}

func (l *LRange) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	start,_ := strconv.Atoi(args[1])
	stop,_ := strconv.Atoi(args[2])
	values := l.Storage.GetSliceFromList(args[0],start, stop) 
	encoded := l.Parser.EncodeAsArray(values)
	_, err := conn.Write([]byte(encoded))
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
