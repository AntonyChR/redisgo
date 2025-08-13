package command

import "context"

type CommandHandler interface {
	Execute(args []string, ctx *context.Context)
}

