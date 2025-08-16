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
	server *network.TcpServer
	parser protocol.Parser
	handlers map[string] command.CommandHandler
	ctx context.Context
}

func (r *Redis) Start(){
	r.server.Start(r.handleConnection)

}

func (r *Redis) handleConnection(conn net.Conn) {
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

		decodedData, _ := r.parser.Decode(data)
		commands,_ := command.ExtractCommandsFromParsedData(decodedData)

		for _, c := range commands {
			if handler, ok := r.handlers[c.Name]; ok {
				resp, _ := handler.Execute(c.Args, &r.ctx)
				conn.Write(resp)
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
	handlers["get"] = &command.GetHandler{Storage: storage}
	handlers["set"] = &command.SetHandler{Storage: storage, ReplicaChan: replicaChan}

	server, _ := network.CreateNewServer("3000", "master", "")

	redis := Redis{
		server: server,
		handlers: handlers,
		parser: p,
		ctx: ctx,
	}

	redis.Start()

}
