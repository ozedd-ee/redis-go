package commands

import (
	"strings"

	"github.com/ozedd-ee/redis-go/serializer"
)

func HandleCommand(c string, *serializer.Serializer) string {
	// split command string into array of the command and its options
	cmdArr := strings.Split(c, " ")
	cmd := cmdArr[0]

	switch cmd {
	case "PING":
		return ping()
	case "ECHO":
		if len(cmdArr) < 2 {
			return s.SerializeSimpleError("err", "No message to ECHO")
		}
		msg := strings.Join(cmdArr[1:], " ")
		return echo(msg)
	default:
		return s.SerializeSimpleError("err", "Invalid command")
	}

}