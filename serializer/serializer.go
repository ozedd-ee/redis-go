package serializer

import "fmt"
import "strings"

const CRLF = "\r\n"

func serializeSimpleString(s string) string {
	return "+" + s + CRLF
}

func serializeSimpleError(prefix string, msg string) string {
	p := strings.ToUpper(prefix)
	err := "-" + p + " " + msg + CRLF
	return err
}

// v: integer value
// s: sign (true for +, false for -) 
func serializeInteger(v int, s bool) string {
	if s {
        return ":" + fmt.Sprint(v) + CRLF
    } else {
        return ":" + "-" + fmt.Sprint(v) + CRLF
    }
}

func serializeBulkString(s string) string {
    l := len(s)
    return "$" + fmt.Sprint(l) + CRLF + s + CRLF
}

func nullBulkString() string {
    return "$-1\r\n"
}
