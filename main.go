package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"

	network "redisgo/network"

	command "redisgo/command"
	protocol "redisgo/protocol"
	storage "redisgo/storage"
)

var SERVER_PORT = flag.String("port", "6379", "Port to listen on")
var REPLICA_OF = flag.String("replicaof", "", "Replicate to another server")

type Redis struct {
	server   *network.TcpServer
	parser   protocol.Parser
	handlers map[string]command.CommandHandler
	ctx      context.Context
}

func (r *Redis) Start() {
	r.server.Start(r.handleConnection)

}

func (r *Redis) handleConnection(conn net.Conn) {
	log.Printf("New connection from %s\n", conn.RemoteAddr().String())
	buff := make([]byte, 1024)
	defer conn.Close()
	for {
		n, err := conn.Read(buff)

		if err == io.EOF {
			log.Println("EOF, connection closed", err)
			return
		}

		if err != nil {
			log.Println("error reading data, ", err)
			return
		}

		if n == 0 {
			log.Println("data length: 0")
			return
		}

		data := buff[:n]

		decodedData, err := r.parser.Decode(data)
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
			if handler, ok := r.handlers[c.Name]; ok {
				resp, err := handler.Execute(c.Args, &r.ctx)
				if err != nil {
					log.Println("error executing command, ", err)
					return
				}
				conn.Write(resp)
			} else {
				log.Printf("Unknown command: %s\n", c.Name)
				conn.Write([]byte("-ERR unknown command '" + c.Name + "'\r\n"))
			}
		}

	}
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage := storage.NewStorage()

	p := &protocol.RedisProtocolParser{}
	replicaChan := make(chan []byte)

	handlers := make(map[string]command.CommandHandler)
	handlers["ping"] = &command.PingHandler{}
	handlers["get"] = &command.GetHandler{Storage: storage}
	handlers["set"] = &command.SetHandler{Storage: storage, ReplicaChan: replicaChan}

	server, _ := network.CreateNewServer("3000", "master", "")

	redis := Redis{
		server:   server,
		handlers: handlers,
		parser:   p,
		ctx:      ctx,
	}

	redis.Start()

}
