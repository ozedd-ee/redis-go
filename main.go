package main

import "github.com/ozedd-ee/redis-go/server"

func main() {
	server.Start(":6379")
}