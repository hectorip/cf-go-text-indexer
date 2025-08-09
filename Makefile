APP := text-indexer
BIN := bin/$(APP)

.PHONY: all build clean cross

all: build

build:
	mkdir -p bin
	go build -o $(BIN) ./...

clean:
	rm -rf bin

cross: clean
	mkdir -p bin
	GOOS=linux   GOARCH=amd64  go build -o bin/$(APP)-linux-amd64 ./...
	GOOS=linux   GOARCH=arm64  go build -o bin/$(APP)-linux-arm64 ./...
	GOOS=darwin  GOARCH=amd64  go build -o bin/$(APP)-darwin-amd64 ./...
	GOOS=darwin  GOARCH=arm64  go build -o bin/$(APP)-darwin-arm64 ./...
	GOOS=windows GOARCH=amd64  go build -o bin/$(APP)-windows-amd64.exe ./...
