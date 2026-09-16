package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"redis-go/commands"
	"redis-go/serializer"
)

type Server struct {
	listener net.Listener
	wg       sync.WaitGroup
}

func New(addr string) (*Server, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Server{listener: l}, nil
}

func (srv *Server) Addr() string {
	return srv.listener.Addr().String()
}

func (srv *Server) Serve() {
	fmt.Println("Server listening on", srv.listener.Addr())
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			break // listener was closed
		}
		srv.wg.Add(1)
		go func() {
			defer srv.wg.Done()
			handleConnection(conn)
		}()
	}
	srv.wg.Wait()
}

func (srv *Server) Stop() {
	srv.listener.Close()
}

// Start is the main entrypoint: listens on addr and handles OS signals for clean shutdown.
func Start(addr string) {
	srv, err := New(addr)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		srv.Stop()
	}()

	srv.Serve()
	fmt.Println("Server stopped")
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		msg, err := readRESP(r)
		if err != nil {
			if err != io.EOF {
				log.Printf("read error: %v", err)
			}
			return
		}
		response := processMessage(msg)
		if _, err := conn.Write([]byte(response)); err != nil {
			log.Printf("write error: %v", err)
			return
		}
	}
}

// readRESP reads exactly one complete RESP array frame from r.
// It handles bulk strings of arbitrary size by reading the declared byte count
// exactly, so large payloads never truncate.
func readRESP(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) == 0 || line[0] != '*' {
		return line, nil
	}

	count, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(line)

	for i := 0; i < count; i++ {
		hdr, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		sb.WriteString(hdr)

		if len(hdr) > 0 && hdr[0] == '$' {
			length, err := strconv.Atoi(strings.TrimSpace(hdr[1:]))
			if err != nil {
				return "", err
			}
			data := make([]byte, length+2) // +2 for trailing CRLF
			if _, err = io.ReadFull(r, data); err != nil {
				return "", err
			}
			sb.Write(data)
		}
	}

	return sb.String(), nil
}

func processMessage(message string) string {
	s := serializer.Serializer{}
	cmdString, err := s.DeserializeMessage(message)
	if err != nil {
		return s.SerializeSimpleError("err", err.Error())
	}
	return commands.HandleCommand(cmdString, &s)
}
