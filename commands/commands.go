package commands

import "github.com/ozedd-ee/redis-go/serializer"

var s = serializer.Serializer{}

func ping() string {
	return s.SerializeSimpleString("PONG")
}

func echo(msg string) string {
	return s.SerializeSimpleString(msg)
}
