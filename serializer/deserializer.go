package serializer

import (
	"log"
	"strconv"
	"strings"
)

// Returning of error types(or any RESP types) should happen at server level, not during de-serialization
func (s *Serializer) deserializeMessage(respArr string) string {
    if string(respArr[0]) != "*" {
		log.Fatal("Message is not a RESP array")
	}
	respArr = respArr[1:]
	// Split at first CRLF to separate array length from elements
	arrLengthString, respArr, found := strings.Cut(respArr, "\r\n")
    // for storing plain command strings
    var plainCmdArr []string;
	if found {
		// Try converting length to integer
		arrLength, err := strconv.Atoi(string(arrLengthString))
		if err != nil {
			log.Fatal("Array length not specified")
		}

		// Split RESP array of bulk strings into Go array of RESP bulk strings (without the identifiers)
		bulkStringArr := strings.Split(respArr, "$")

        // Check to ensure that the specified length matches the actual length
        if (len(bulkStringArr) != arrLength) {
            log.Fatal("Specified length does not match actual length")
        }

		for _, v := range bulkStringArr {
            // extract length and remove starting CRLF
            _, noPrefixV, _ := strings.Cut(v, "\r\n")
            // remove ending CRLF
            noSuffixV, s2 := strings.CutSuffix(noPrefixV, "\r\n")
            if !s2 {
                log.Fatal("No closing CRLF for bulk string")
            }
            // append cleaned string without CRLF to array
            plainCmdArr = append(plainCmdArr, noSuffixV)
		}
	}
    return strings.Join(plainCmdArr, "")
}
