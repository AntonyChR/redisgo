package main

import (
	"context"
	"flag"
	command "redisgo/command"
	network "redisgo/network"
	utils "redisgo/utils"
	protocol "redisgo/protocol"
	redis "redisgo/redis"
	storage "redisgo/storage"
)

var SERVER_PORT = flag.String("port", "6379", "Port to listen on")
var REPLICA_OF = flag.String("replicaof", "", "Replicate to another server")

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage := storage.NewStorage()

	p := &protocol.RedisProtocolParser{}
	replicaChan := make(chan []byte)

	handlers := make(map[string]command.CommandHandler)
	handlers["ping"] = &command.PingHandler{}
	handlers["echo"] = &command.EchoHandler{Parser: p}
	handlers["get"] = &command.GetHandler{Storage: storage, Parser: p}
	handlers["set"] = &command.SetHandler{Storage: storage, ReplicaChan: replicaChan}
	handlers["rpush"] = &command.RPush{Storage: storage}

	server, _ := network.CreateNewServer(*SERVER_PORT, "master", "")

	instanceInfo := redis.InstanceInfo{
		Port: *SERVER_PORT,
		Id: utils.GenerateUUID(),
		Offset: 0, 
	}

	redis := redis.Redis{
		Server:   server,
		Handlers: handlers,
		Parser:   p,
		Info:     &instanceInfo,
		Ctx:      ctx,
	}

	redis.Start()

}
