package serializer

import "fmt"
import "strings"

const CRLF = "\r\n"

type doubleOpts struct {
    isNegative bool; // true for +, false for -
    integer int;
    fraction int;
    hasPosExponent bool; 
    hasNegExponent bool; // Has negative exponent
    exponent int;
}

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

func serializeBool(b bool) string {
	if b {
        return "#" + "t" + CRLF
    } else {
        return "#" + "f"  + CRLF
    }
}

func serializeDouble(params doubleOpts) string {
    respDouble := ","
    if params.isNegative {
        respDouble += "-" + fmt.Sprint(params.integer)
    } else {
        respDouble += fmt.Sprint(params.integer)
    }
    if params.fraction != 0 {
        respDouble += "." + fmt.Sprint(params.fraction)
    }
    if params.hasPosExponent {
        respDouble += "e" + fmt.Sprint(params.exponent)
    }
    if params.hasNegExponent {
        respDouble += "e" + "-" + fmt.Sprint(params.exponent)
    }
    return respDouble
}

func null() string {
    return "_" + CRLF
}

func nullBulkString() string {
    return "$-1\r\n"
}
