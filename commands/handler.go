package commands

import (
	"strings"

	"github.com/ozedd-ee/redis-go/serializer"
)

func HandleCommand(c string, s *serializer.Serializer) string {
	// split command string into array of the command and its options using CRLF separator retained during deserialization
	cmdArr := strings.Split(c, "\r\n")
	cmdArr = cmdArr[:len(cmdArr)-1]
	cmd := strings.ToUpper(cmdArr[0])

	switch cmd {
	case "PING":
		return ping()
	case "ECHO":
		if len(cmdArr) < 2 {
			return s.SerializeSimpleError("err", "No message to ECHO")
		}
		msg := cmdArr[1]
		return echo(msg)
	case "SET":
		if len(cmdArr) < 2 {
			return s.SerializeSimpleError("err", "No key and value specified")
		}
		if len(cmdArr) < 3 {
			return s.SerializeSimpleError("err", "No value specified for key")
		}
		key, val := cmdArr[1], cmdArr[2]
		return set(key, val, cmdArr[3:]...)
	case "GET":
		if len(cmdArr) < 2 {
			return s.SerializeSimpleError("err", "No key specified")
		}
		return get(cmdArr[1])
	case "INFO":
		return s.SerializeBulkString("# Server \r\n redis_version: 5 \r\n tcp_port:6379")
	default:
		return s.SerializeSimpleError("err", "Invalid command")
	}

}
