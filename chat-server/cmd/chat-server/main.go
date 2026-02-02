package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chat-server/internal/config"
	"chat-server/internal/hub"
	"chat-server/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "chat-server: ", log.LstdFlags|log.LUTC|log.Lmsgprefix)

	cfg, err := config.FromEnv()
	if err != nil {
		logger.Fatalf("failed to load config: %v", err)
	}

	tcpListener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		logger.Fatalf("failed to listen on %q: %v", cfg.ListenAddr, err)
	}
	// Best-effort cleanup in case we exit due to a fatal error.
	defer func() {
		_ = tcpListener.Close()
	}()

	rootContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	chatHub := hub.New(logger, cfg)
	tcpServer := server.NewTCPServer(logger, cfg, chatHub)

	go func() {
		<-rootContext.Done()

		const shutdownTimeout = 5 * time.Second
		shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if shutdownErr := tcpServer.Shutdown(shutdownContext); shutdownErr != nil {
			logger.Printf("shutdown error: %v", shutdownErr)
		}
	}()

	logger.Printf("listening on %s", cfg.ListenAddr)

	serveErr := tcpServer.Serve(rootContext, tcpListener)
	if serveErr == nil {
		logger.Printf("server stopped")
		return
	}

	if errors.Is(serveErr, net.ErrClosed) || errors.Is(rootContext.Err(), context.Canceled) {
		logger.Printf("server stopped")
		return
	}

	logger.Fatalf("server error: %v", serveErr)
}
