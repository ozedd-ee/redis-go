package serializer

import (
	"errors"
	"strconv"
	"strings"
)

func (s *Serializer) DeserializeMessage(respArr string) (string, error) {
	if !strings.HasPrefix(respArr, "*") {
		return "", errors.New("message is not a RESP array")
	}

	lines := strings.Split(respArr, CRLF)
	if len(lines) < 2 {
		return "", errors.New("malformed input")
	}

	// Try converting length to integer
	_, err := strconv.Atoi(lines[0][1:])
	if err != nil {
		return "", errors.New("invalid array length")
	}

	var result strings.Builder
	for i := 1; i < len(lines)-1; i++ {
		if strings.HasPrefix(lines[i], "$") && i+1 < len(lines) {
			result.WriteString(lines[i+1])
			result.WriteString(" ") // Add space between command and arguments
			i++ // Skip the next line as it is the bulk string content
		}
	}

	return result.String(), nil
}
