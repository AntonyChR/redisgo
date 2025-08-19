package command

import (
	"context"
	"errors"
	"net"
	"redisgo/protocol"
	storage "redisgo/storage"
)

// PING
type PingHandler struct{}

func (p *PingHandler) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	_, err := conn.Write([]byte("+PONG\r\n"))
	return err
}

// ECHO
type EchoHandler struct{}

func (e *EchoHandler) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	if len(args) == 0 {
		return errors.New("ECHO command requires an argument")
	}
	_, err := conn.Write([]byte(args[0]))
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
	s.Storage.Set(args[0], args[1])
	// TODO: implement expiration logic
	_, err := conn.Write(okResponse())
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
