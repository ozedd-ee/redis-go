package serializer

import (
	"fmt"
	"strings"
)

const CRLF = "\r\n"

type doubleOpts struct {
	isNegative     bool // true for +, false for -
	integer        int
	fraction       int
	hasPosExponent bool
	hasNegExponent bool // Has negative exponent
	exponent       int
}

type Serializer struct {

}

func (s *Serializer) SerializeSimpleString(msg string) string {
	return "+" + msg + CRLF
}

func (s *Serializer) SerializeSimpleError(prefix string, msg string) string {
	p := strings.ToUpper(prefix)
	err := "-" + p + " " + msg + CRLF
	return err
}

// v: integer value
// isPos: (true for +, false for -)
func (s *Serializer) SerializeInteger(v int, isPos bool) string {
	if isPos {
		return ":" + fmt.Sprint(v) + CRLF
	} else {
		return ":" + "-" + fmt.Sprint(v) + CRLF
	}
}

func (s *Serializer) SerializeBulkString(msg string) string {
	l := len(msg)
	return "$" + fmt.Sprint(l) + CRLF + msg + CRLF
}

func (s *Serializer) SerializeBool(b bool) string {
	if b {
		return "#" + "t" + CRLF
	} else {
		return "#" + "f" + CRLF
	}
}

func (s *Serializer) SerializeDouble(params doubleOpts) string {
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

func (s *Serializer) SerializeBigNumber(bigNum string, isNegative bool) string {
	if isNegative {
		return "(" + "-" + bigNum + CRLF
	} else {
		return "(" + bigNum + CRLF
	}
}

func (s *Serializer) SerializeBulkError(e string) string {
	l := len(e)
	return "!" + fmt.Sprint(l) + CRLF + e + CRLF
}

func (s *Serializer) Null() string {
	return "_" + CRLF
}

func (s *Serializer) NullBulkString() string {
	return "$-1\r\n"
}

func (s *Serializer) SerializeArray(elements ...string) string {
	respArray := "*"
	arrLength := len(elements)
	respArray += fmt.Sprint(arrLength) + CRLF
	for _, v := range elements {
		respArray += s.SerializeBulkString(v)
	}
	return respArray
}
