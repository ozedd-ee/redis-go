package serializer

import (
	"fmt"
	"strconv"
	"strings"
)

const CRLF = "\r\n"

type doubleOpts struct {
	isNegative     bool // true for -, false for +
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
		respDouble += "-" + strconv.Itoa(params.integer)
	} else {
		respDouble += strconv.Itoa(params.integer)
	}
	if params.fraction != 0 {
		respDouble += "." + strconv.Itoa(params.fraction)
	}
	if params.hasPosExponent {
		respDouble += "e" + strconv.Itoa(params.exponent)
	} else if params.hasNegExponent {
		respDouble += "e" + "-" + strconv.Itoa(params.exponent)
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
	var b strings.Builder
	b.WriteString("*")
	b.WriteString(strconv.Itoa(len(elements)))
	b.WriteString(CRLF)

	for _, v := range elements {
		b.WriteString(s.SerializeBulkString(v))
	}
	return b.String()
}
