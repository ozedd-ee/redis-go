package serializer

import (
	"log"
	"strconv"
	"strings"
)

func (s *Serializer) DeserializeMessage(respArr string) string {
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
            // Remove length and following CRLF
            _, noPrefixV, success := strings.Cut(v, "\r\n")
			if !success {
                log.Fatal("No opening CRLF for bulk string")
            }
            // append cleaned string to array (Retain closing CRLF as separator)
            plainCmdArr = append(plainCmdArr, noPrefixV)
		}
	}
    return strings.Join(plainCmdArr, "")
}
