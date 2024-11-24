package commands

import (
	"strings"

	"github.com/ozedd-ee/redis-go/serializer"
)

func HandleCommand(c string, s *serializer.Serializer) string {
	// split command string into array of the command and its options using CRLF separator retained during deserialization
	cmdArr := strings.Split(c, "\r\n")
	cmd := cmdArr[0]

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
		if len(cmdArr) < 3 {
			return s.SerializeSimpleError("err", "No value specified for key")
		}
		key, val := cmdArr[1], cmdArr[2]
		return set(key, val)
	case "GET":
		if len(cmdArr) < 2 {
			return s.SerializeSimpleError("err", "No key specified")
		}
		return get(cmdArr[1])
	default:
		return s.SerializeSimpleError("err", "Invalid command")
	}

}
