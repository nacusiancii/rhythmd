package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	FocusDuration = 25 * time.Minute
	RelaxDuration = 5 * time.Minute
	SocketPath    = "/tmp/hourglass.sock"
)

type State struct {
	Mode      string
	StartTime time.Time
	Duration  int
}

type Service struct {
	state     State
	listeners []net.Conn
}

func (s *Service) startCycle(mode string) {
	var duration time.Duration
	if mode == "focus" {
		duration = FocusDuration
	} else {
		duration = RelaxDuration
	}

	s.state = State{
		Mode:      mode,
		StartTime: time.Now(),
		Duration:  int(duration.Seconds()),
	}

	log.Printf("Starting %s cycle (%v)", mode, duration)

	time.Sleep(duration)

	if mode == "focus" {
		s.startCycle("relax")
	} else {
		s.startCycle("focus")
	}
}

func (s *Service) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Send current state immediately
	data, _ := json.Marshal(s.state)
	log.Println("Data:", data)
	conn.Write(append(data, '\n'))

	// Add to listeners for future updates
	s.listeners = append(s.listeners, conn)

	// Keep connection alive
	buf := make([]byte, 1024)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			return
		}
	}
}

func (s *Service) startSocket() {
	os.Remove(SocketPath)

	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		log.Fatal("Failed to create socket:", err)
	}
	defer listener.Close()

	log.Printf("Socket listening on %s", SocketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
		}
		go s.handleConnection(conn)
	}
}

func main() {
	log.SetFlags(log.Ltime)
	log.Println("Hourglass service starting...")

	service := &Service{}

	log.Println(service)

	// Start unix socket for IPC
	go service.startSocket()
	log.Println("Started Socket")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		os.Remove(SocketPath)
		os.Exit(0)
	}()

	service.startCycle("focus")
}
