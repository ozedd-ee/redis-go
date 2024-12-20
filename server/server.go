package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/ozedd-ee/redis-go/commands"
	"github.com/ozedd-ee/redis-go/serializer"
)

func Start(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()
	fmt.Println("Server is listening on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection", err)
			continue
		}
        fmt.Println("Connection Established")
		// Handle connection in separate go routine to allow server to serve multiple clients concurrently
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
        reader := bufio.NewReader(conn)
        var message []byte
        buffer := make([]byte, 1024000) // 1 MB buffer
		// Use buffer to read entire message
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				// End-of-stream
				break
			}
			fmt.Println("Error reading message: ", err)
			return
		}
		// Append read data to message
		message = append(message, buffer[:n]...)  
        
        // Process message and respond
        response := processMessage(string(message))
        _, writerErr := conn.Write([]byte(response))
        if writerErr != nil {
            fmt.Println("Error writing response:", err)
        }
	}
}

func processMessage(message string) string {
	// Initialize serializer
	s := serializer.Serializer{}

	cmdString, err := s.DeserializeMessage(message)
    if err != nil {
        return s.SerializeSimpleError("err", err.Error())
    }
	res := commands.HandleCommand(cmdString, &s)
    return res
}
