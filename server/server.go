package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
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
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("read error: %v", err)
			}
			return
		}
		response := processMessage(string(buf[:n]))
		if _, err := conn.Write([]byte(response)); err != nil {
			log.Printf("write error: %v", err)
			return
		}
	}
}

func processMessage(message string) string {
	s := serializer.Serializer{}
	cmdString, err := s.DeserializeMessage(message)
	if err != nil {
		return s.SerializeSimpleError("err", err.Error())
	}
	return commands.HandleCommand(cmdString, &s)
}
