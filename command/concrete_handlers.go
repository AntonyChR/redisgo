package command

import (
	"context"
	"errors"
	"redisgo/protocol"
	storage "redisgo/storage"
)

// PING
type PingHandler struct{}

func (p *PingHandler) Execute(args []string, ctx *context.Context) ([]byte, error) {
	return []byte("+PONG\r\n"), nil
}

// ECHO
type EchoHandler struct{}

func (e *EchoHandler) Execute(args []string, ctx *context.Context) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("ECHO command requires an argument")
	}
	return []byte(args[0]), nil
}

// GET
type GetHandler struct {
	Storage *storage.Storage
	Parser  protocol.Parser
}

func (g *GetHandler) Execute(args []string, ctx *context.Context) ([]byte, error) {
	value, ok := g.Storage.Get(args[0])
	if !ok {
		return nilResponse(), nil // Return null bulk string for non-existing key
	}
	encondedResp := g.Parser.EncodeBulkString(value, true)
	return []byte(encondedResp), nil
}

// SET
type SetHandler struct {
	Storage     *storage.Storage
	ReplicaChan chan []byte
}

func (s *SetHandler) Execute(args []string, ctx *context.Context) ([]byte, error) {
	if len(args[0]) == 0 {
		return nil, errors.New("invalid key value")
	}
	s.Storage.Set(args[0], args[1])
	// TODO: implement expiration logic
	return okResponse(), nil
}

func okResponse() []byte {
	return []byte("+OK\r\n")
}

func nilResponse() []byte {
	return []byte("$-1\r\n")
}
