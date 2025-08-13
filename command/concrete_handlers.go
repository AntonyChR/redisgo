package command

import (
	"context"
	storage "redisgo/storage"
)

type PingHandler struct{}

func (p *PingHandler) Execute(ctx *context.Context){
	//ctx.write("pong")	
}

type GetHandler struct{
	Storage *storage.Storage
}

func (g *GetHandler) Execute( args []string, ctx *context.Context){
	g.Storage.Get("")
}


type SetHandler struct{
	Storage *storage.Storage
}

func (s *SetHandler) Execute( args []string, ctx *context.Context){
	s.Storage.Set("","")
}

