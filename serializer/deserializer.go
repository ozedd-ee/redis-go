package serializer

import (
	"errors"
	"strconv"
	"strings"
)

func (s *Serializer) DeserializeMessage(respArr string) (string, error) {
	if string(respArr[0]) != "*" {
		return "", errors.New("message is not a RESP array")
	}
	respArr = respArr[1:]
	// Split at first CRLF to separate array length from elements
	arrLengthString, respArr, found := strings.Cut(respArr, "\r\n")
	// for storing plain command strings
	var plainCmdArr []string
	if found {
		// Try converting length to integer
		_, err := strconv.Atoi(string(arrLengthString))
		if err != nil {
			return "", errors.New("array length not specified")
		}

		// Split RESP array of bulk strings into Go array of RESP bulk strings (without the identifiers)
		_, bulkSA, _ := strings.Cut(respArr, "$")
		bulkStringArr := strings.Split(bulkSA, "$")

		for _, v := range bulkStringArr {
			// Remove length and following CRLF
			_, noPrefixV, success := strings.Cut(v, "\r\n")
			if !success {
				return "", errors.New("no opening CRLF for bulk string")
			}
			// append cleaned string to array (Retain closing CRLF as separator)
			plainCmdArr = append(plainCmdArr, noPrefixV)
		}
	}
	return strings.Join(plainCmdArr, ""), nil
}
