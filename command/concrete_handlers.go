package command

import (
	"context"
	"errors"
	"redisgo/protocol"
	storage "redisgo/storage"
)

type PingHandler struct{}

func (p *PingHandler) Execute(ctx *context.Context) ([]byte, error) {
	return okResponse(), nil
}

type GetHandler struct {
	Storage *storage.Storage
	parser  protocol.Parser
}

func (g *GetHandler) Execute(args []string, ctx *context.Context) ([]byte, error) {
	value, _ := g.Storage.Get(args[0])
	encondedResp := g.parser.EncodeBulkString(value, true)
	return []byte(encondedResp), nil
}

type SetHandler struct {
	Storage     *storage.Storage
	ReplicaChan chan []byte
}

func (s *SetHandler) Execute(args []string, ctx *context.Context) ([]byte, error) {
	if len(args[0]) == 0 {
		return nil, errors.New("invalid key value")
	}
	s.Storage.Set(args[0], args[1])
	return nil, nil
}

func okResponse() []byte {
	return []byte("+OK\r\n")
}
