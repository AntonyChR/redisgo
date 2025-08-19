package command

import (
	"context"
	"net"
)

type CommandHandler interface {
	Execute(args []string, ctx *context.Context, conn net.Conn) error
}
