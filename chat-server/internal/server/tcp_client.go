package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"chat-server/internal/config"
	"chat-server/internal/framing"
	"chat-server/internal/hub"
)

// TCPClient represents a single TCP-connected client.
// It bridges raw network I/O with the hub event model.
type TCPClient struct {
	logger *log.Logger
	cfg    config.Config
	hub    *hub.Hub

	conn     net.Conn
	clientID hub.ClientID

	writeQueue chan []byte

	closeOnce sync.Once
}

// NewTCPClient constructs a TCPClient bound to an existing TCP connection.
func NewTCPClient(
	logger *log.Logger,
	cfg config.Config,
	hubInstance *hub.Hub,
	conn net.Conn,
) *TCPClient {
	clientID := hub.ClientID(fmt.Sprintf(
		"%s->%s",
		conn.RemoteAddr().String(),
		conn.LocalAddr().String(),
	))

	return &TCPClient{
		logger:     logger,
		cfg:        cfg,
		hub:        hubInstance,
		conn:       conn,
		clientID:   clientID,
		writeQueue: make(chan []byte, cfg.WriteQueueDepth),
	}
}

// Run starts the client read/write loops and blocks until the client terminates.
func (c *TCPClient) Run(parentCtx context.Context) {
	c.hub.Register(c.clientID, c)

	clientContext, cancel := context.WithCancel(parentCtx)
	defer cancel()

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()
		c.readLoop(clientContext)
	}()

	go func() {
		defer waitGroup.Done()
		c.writeLoop(clientContext)
	}()

	waitGroup.Wait()

	c.hub.Unregister(c.clientID, "connection closed")
	_ = c.Close()
}

// readLoop reads newline-delimited frames from the TCP connection
// and forwards them to the hub.
func (c *TCPClient) readLoop(ctx context.Context) {
	lineReader := framing.NewLineReader(c.conn, c.cfg.MaxFrameBytes)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if c.cfg.ReadTimeoutSecs > 0 {
			_ = c.conn.SetReadDeadline(
				time.Now().Add(time.Duration(c.cfg.ReadTimeoutSecs) * time.Second),
			)
		}

		frame, err := lineReader.ReadFrame()
		if err != nil {
			c.hub.Unregister(c.clientID, fmt.Sprintf("read error: %v", err))
			return
		}

		c.hub.Deliver(c.clientID, frame)
	}
}

// writeLoop writes outbound frames to the TCP connection.
func (c *TCPClient) writeLoop(ctx context.Context) {
	lineWriter := framing.NewLineWriter(c.conn)

	for {
		select {
		case <-ctx.Done():
			return

		case frame, ok := <-c.writeQueue:
			if !ok {
				return
			}

			writeContext := ctx
			var cancel context.CancelFunc

			if c.cfg.WriteTimeoutSecs > 0 {
				writeContext, cancel = context.WithTimeout(
					ctx,
					time.Duration(c.cfg.WriteTimeoutSecs)*time.Second,
				)
			}

			err := lineWriter.WriteFrame(writeContext, frame)

			if cancel != nil {
				cancel() // cancel immediately; do NOT defer inside the loop
			}

			if err != nil {
				c.hub.Unregister(c.clientID, fmt.Sprintf("write error: %v", err))
				return
			}
		}
	}
}

// Send enqueues a frame for delivery to the client.
func (c *TCPClient) Send(ctx context.Context, frame []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.writeQueue <- frame:
		return nil
	default:
		// Backpressure: if the client is not reading fast enough,
		// fail closed to protect server resources.
		return fmt.Errorf("client write queue is full")
	}
}

// Close closes the client connection and releases resources.
func (c *TCPClient) Close() error {
	var closeError error

	c.closeOnce.Do(func() {
		close(c.writeQueue)
		closeError = c.conn.Close()
	})

	return closeError
}
