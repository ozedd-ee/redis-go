package main

import "redis-go/server"

func main() {
	server.Start(":6379")
}