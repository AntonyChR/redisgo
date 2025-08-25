package redis

import (
	"context"
	"io"
	"log"
	"net"
	command "redisgo/command"
	network "redisgo/network"
	protocol "redisgo/protocol"
)

type Redis struct {
	Info     *InstanceInfo
	Server   *network.TcpServer
	Parser   protocol.Parser
	Handlers map[string]command.CommandHandler
	Ctx      context.Context
}

func (r *Redis) Start() {
	r.Server.Start(r.handleConnection)
}

func (r *Redis) handleConnection(conn net.Conn) {
	buff := make([]byte, 1024)
	defer conn.Close()
	for {
		n, err := conn.Read(buff)

		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("error reading data, ", err)
			return
		}

		if n == 0 {
			log.Println("data length: 0")
			return
		}

		data := buff[:n]

		decodedData, err := r.Parser.Decode(data)
		if err != nil {
			log.Println("error decoding data, ", err)
			return
		}
		commands, err := command.ExtractCommandsFromParsedData(decodedData)
		if err != nil {
			log.Println("error extracting commands, ", err)
			return
		}

		for _, c := range commands {
			log.Printf("Received command: %s with args: %v\n", c.Name, c.Args)
			if handler, ok := r.Handlers[c.Name]; ok {
				err := handler.Execute(c.Args, &r.Ctx, conn)
				if err != nil {
					log.Println("error executing command, ", err)
					return
				}

			} else {
				log.Printf("Unknown command: %s\n", c.Name)
				conn.Write([]byte("-ERR unknown command '" + c.Name + "'\r\n"))
			}
		}

	}
}
