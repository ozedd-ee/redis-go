package commands

import (
	"strings"

	"redis-go/serializer"
)

var commandHandlers = map[string]func([]string, *serializer.Serializer) string{
	"ping":   handlePing,
	"echo":   handleEcho,
	"set":    handleSet,
	"get":    handleGet,
	"info":   handleInfo,
	"exists": handleExists,
	"del":    handleDel,
	"incr":   handleIncr,
	"decr":   handleDecr,
	"lpush":  handleLpush,
	"rpush":  handleRpush,
	"lrange": handleLrange,
}

func HandleCommand(c string, s *serializer.Serializer) string {
	cmdArr := strings.Split(c, " ")
	if len(cmdArr) > 0 && cmdArr[len(cmdArr)-1] == "" {
		cmdArr = cmdArr[:len(cmdArr)-1]
	}
	if len(cmdArr) == 0 {
		return s.SerializeSimpleError("err", "empty command")
	}

	cmd := strings.ToLower(cmdArr[0])
	handler, ok := commandHandlers[cmd]
	if !ok {
		return s.SerializeSimpleError("err", "invalid command")
	}
	return handler(cmdArr, s)
}

func handlePing(args []string, s *serializer.Serializer) string {
	return ping()
}

func handleEcho(args []string, s *serializer.Serializer) string {
	if len(args) < 2 {
		return s.SerializeSimpleError("err", "No message to ECHO")
	}
	return echo(args[1])
}

func handleSet(args []string, s *serializer.Serializer) string {
	if len(args) < 3 {
		return s.SerializeSimpleError("err", "SET requires a key and value")
	}
	return set(args[1], args[2], args[3:]...)
}

func handleGet(args []string, s *serializer.Serializer) string {
	if len(args) < 2 {
		return s.SerializeSimpleError("err", "GET requires a key")
	}
	return get(args[1])
}

func handleInfo(args []string, s *serializer.Serializer) string {
	return s.SerializeBulkString("# Server \r\n redis_version: 5 \r\n tcp_port:6379")
}

func handleExists(args []string, s *serializer.Serializer) string {
	if len(args) < 2 {
		return s.SerializeSimpleError("err", "EXISTS requires at least one key")
	}
	return exists(args[1:]...)
}

func handleDel(args []string, s *serializer.Serializer) string {
	if len(args) < 2 {
		return s.SerializeSimpleError("err", "DEL requires at least one key")
	}
	return deleteKey(args[1:]...)
}

func handleIncr(args []string, s *serializer.Serializer) string {
	if len(args) < 2 {
		return s.SerializeSimpleError("err", "INCR requires a key")
	}
	return increment(args[1])
}

func handleDecr(args []string, s *serializer.Serializer) string {
	if len(args) < 2 {
		return s.SerializeSimpleError("err", "DECR requires a key")
	}
	return decrement(args[1])
}

func handleLpush(args []string, s *serializer.Serializer) string {
	if len(args) < 3 {
		return s.SerializeSimpleError("err", "LPUSH requires a key and at least one value")
	}
	return lpush(args[1], args[2:]...)
}

func handleRpush(args []string, s *serializer.Serializer) string {
	if len(args) < 3 {
		return s.SerializeSimpleError("err", "RPUSH requires a key and at least one value")
	}
	return rpush(args[1], args[2:]...)
}

func handleLrange(args []string, s *serializer.Serializer) string {
	if len(args) < 4 {
		return s.SerializeSimpleError("err", "LRANGE requires key, start, and stop")
	}
	return lrange(args[1], args[2], args[3])
}
