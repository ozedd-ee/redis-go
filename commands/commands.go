package commands

import "github.com/ozedd-ee/redis-go/serializer"

var s = serializer.Serializer{}

var store map[string]string

func ping() string {
	return s.SerializeSimpleString("PONG")
}

func echo(msg string) string {
	return s.SerializeSimpleString(msg)
}

func set(key string, val string) string {
	store[key] = val
	return s.SerializeSimpleString("OK")
}

func get(key string) string {
	val, ok := store[key] 
	if !ok {
		return s.Null()
	}
	return s.SerializeBulkString(val)
}
