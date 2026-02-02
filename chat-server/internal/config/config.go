package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ListenAddr        string
	MaxFrameBytes     int
	WriteQueueDepth   int
	ReadTimeoutSecs   int
	WriteTimeoutSecs  int
	IdleTimeoutSecs   int
	MaxUsernameLength int
	MaxRoomNameLength int
}

func FromEnv() (Config, error) {
	const (
		defaultListenAddr      = ":8080"
		defaultMaxFrameBytes   = 64 * 1024
		defaultWriteQueueDepth = 128

		defaultReadTimeoutSecs  = 0
		defaultWriteTimeoutSecs = 0
		defaultIdleTimeoutSecs  = 0

		protocolMaxUsernameLength = 8
		protocolMaxRoomNameLength = 16
	)

	listenAddr := getEnvString("CHAT_SERVER_ADDR", defaultListenAddr)

	maxFrameBytes, err := getEnvIntStrict("CHAT_SERVER_MAX_FRAME_BYTES", defaultMaxFrameBytes)
	if err != nil {
		return Config{}, err
	}
	writeQueueDepth, err := getEnvIntStrict("CHAT_SERVER_WRITE_QUEUE_DEPTH", defaultWriteQueueDepth)
	if err != nil {
		return Config{}, err
	}
	readTimeoutSecs, err := getEnvIntStrict("CHAT_SERVER_READ_TIMEOUT_SECS", defaultReadTimeoutSecs)
	if err != nil {
		return Config{}, err
	}
	writeTimeoutSecs, err := getEnvIntStrict("CHAT_SERVER_WRITE_TIMEOUT_SECS", defaultWriteTimeoutSecs)
	if err != nil {
		return Config{}, err
	}
	idleTimeoutSecs, err := getEnvIntStrict("CHAT_SERVER_IDLE_TIMEOUT_SECS", defaultIdleTimeoutSecs)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		ListenAddr:        listenAddr,
		MaxFrameBytes:     maxFrameBytes,
		WriteQueueDepth:   writeQueueDepth,
		ReadTimeoutSecs:   readTimeoutSecs,
		WriteTimeoutSecs:  writeTimeoutSecs,
		IdleTimeoutSecs:   idleTimeoutSecs,
		MaxUsernameLength: protocolMaxUsernameLength,
		MaxRoomNameLength: protocolMaxRoomNameLength,
	}

	if cfg.MaxFrameBytes <= 0 {
		return Config{}, fmt.Errorf("invalid CHAT_SERVER_MAX_FRAME_BYTES: %d", cfg.MaxFrameBytes)
	}
	if cfg.WriteQueueDepth <= 0 {
		return Config{}, fmt.Errorf("invalid CHAT_SERVER_WRITE_QUEUE_DEPTH: %d", cfg.WriteQueueDepth)
	}
	if cfg.ReadTimeoutSecs < 0 {
		return Config{}, fmt.Errorf("invalid CHAT_SERVER_READ_TIMEOUT_SECS: %d", cfg.ReadTimeoutSecs)
	}
	if cfg.WriteTimeoutSecs < 0 {
		return Config{}, fmt.Errorf("invalid CHAT_SERVER_WRITE_TIMEOUT_SECS: %d", cfg.WriteTimeoutSecs)
	}
	if cfg.IdleTimeoutSecs < 0 {
		return Config{}, fmt.Errorf("invalid CHAT_SERVER_IDLE_TIMEOUT_SECS: %d", cfg.IdleTimeoutSecs)
	}

	return cfg, nil
}

func getEnvString(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntStrict(key string, defaultValue int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s=%q: %w", key, value, err)
	}
	return parsed, nil
}
