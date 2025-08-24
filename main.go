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
	handlers[protocol.PING] = &command.PingHandler{}
	handlers[protocol.ECHO] = &command.EchoHandler{Parser: p}
	handlers[protocol.GET] = &command.GetHandler{Storage: storage, Parser: p}
	handlers[protocol.SET] = &command.SetHandler{Storage: storage, ReplicaChan: replicaChan}
	handlers[protocol.RPUSH] = &command.RPush{Storage: storage}
	handlers[protocol.LRANGE] = &command.LRange{Storage: storage, Parser: p}
	handlers[protocol.LPUSH] = &command.LPush{Storage: storage}
	handlers[protocol.LLEN] = &command.LLEN{Storage: storage}
	handlers[protocol.LPOP] = &command.LPOP{Storage: storage, Parser: p}
	handlers[protocol.BLPOP] = &command.BLPOP{Storage: storage, Parser: p}
	handlers[protocol.TYPE] = &command.Type{Storage: storage, Parser: p}
	handlers[protocol.XADD] = &command.XAdd{Storage: storage, Parser: p}

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
