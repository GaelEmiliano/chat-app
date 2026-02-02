BINARY_NAME := chat-server
CMD_PATH := ./cmd/chat-server

.PHONY: build run test fmt vet lint clean

build:
	go build -o bin/$(BINARY_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH)

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

lint: fmt vet

clean:
	rm -rf bin
