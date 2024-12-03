package commands

import (
	"strings"

	"github.com/ozedd-ee/redis-go/serializer"
)

var s = serializer.Serializer{}

var store = make(map[string]Value)

func ping() string {
	return s.SerializeSimpleString("PONG")
}

func echo(msg string) string {
	return s.SerializeSimpleString(msg)
}

func set(key string, val string, opts ...string) string {
	if len(opts) == 0 {
		exp := Expiry{option: NONE, time: 0}
		value := Value{expiry: exp, value: val}
		store[key] = value
		return s.SerializeSimpleString("OK")
	}
	for i, v := range opts {
		V := strings.ToUpper(v)
		switch V {
		case "EX":
			return handleSet(key, val, EX, opts[i+1])
		case "PX":
			return handleSet(key, val, PX, opts[i+1])
		case "EXAT":
			return handleSet(key, val, EXAT, opts[i+1])
		case "PXAT":
			return handleSet(key, val, PXAT, opts[i+1])
		}
	}
	return s.SerializeSimpleError("err", "Invalid Expiry option specified")
}

// TODO: Update GET, Work on Expire func, CheckExpiry
func get(key string) string {
	val, ok := store[key]
	if !ok {
		return s.Null()
	}
	return s.SerializeBulkString(val.value)
}
