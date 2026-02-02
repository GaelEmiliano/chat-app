package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"chat-server/internal/config"
	"chat-server/internal/hub"
)

// TCPServer accepts TCP connections and wires them to the Hub.
// It owns the listener lifecycle and coordinates graceful shutdown.
type TCPServer struct {
	logger *log.Logger
	cfg    config.Config
	hub    *hub.Hub

	listenerMu sync.Mutex
	listener   net.Listener

	clientsWaitGroup sync.WaitGroup
	hubWaitGroup     sync.WaitGroup
}

// NewTCPServer creates a new TCPServer instance.
// The Hub must already be constructed and will be run by Serve.
func NewTCPServer(
	logger *log.Logger,
	cfg config.Config,
	hubInstance *hub.Hub,
) *TCPServer {
	return &TCPServer{
		logger: logger,
		cfg:    cfg,
		hub:    hubInstance,
	}
}

// Serve starts accepting connections and blocks until the server stops.
// It returns net.ErrClosed on normal shutdown.
func (s *TCPServer) Serve(ctx context.Context, listener net.Listener) error {
	s.listenerMu.Lock()
	s.listener = listener
	s.listenerMu.Unlock()

	s.hubWaitGroup.Add(1)
	go func() {
		defer s.hubWaitGroup.Done()
		s.hub.Run(ctx)
	}()

	for {
		connection, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return net.ErrClosed
			}

			select {
			case <-ctx.Done():
				return net.ErrClosed
			default:
				return fmt.Errorf("accept connection: %w", err)
			}
		}

		s.clientsWaitGroup.Add(1)
		go func(conn net.Conn) {
			defer s.clientsWaitGroup.Done()
			client := NewTCPClient(s.logger, s.cfg, s.hub, conn)
			client.Run(ctx)
		}(connection)
	}
}

// Shutdown gracefully stops accepting new connections and waits for
// all clients and the hub to terminate.
func (s *TCPServer) Shutdown(ctx context.Context) error {
	s.listenerMu.Lock()
	listener := s.listener
	s.listenerMu.Unlock()

	if listener != nil {
		_ = listener.Close()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.clientsWaitGroup.Wait()
		s.hubWaitGroup.Wait()
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	case <-done:
		return nil
	}
}

// withOptionalDeadline returns a context with a deadline applied
// only if seconds is greater than zero.
func withOptionalDeadline(parent context.Context, seconds int) context.Context {
	if seconds <= 0 {
		return parent
	}

	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	ctx, _ := context.WithDeadline(parent, deadline)
	return ctx
}
