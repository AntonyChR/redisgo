package main

import (
	"flag"
)

var SERVER_PORT = flag.String("port", "6379", "Port to listen on")
var REPLICA_OF = flag.String("replicaof", "", "Replicate to another server")

func main() {
	flag.Parse()
	println("RedisGo Server started")

}
