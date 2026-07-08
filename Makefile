BINARY=arduino-mcp
BUILD_DIR=.
CMD_DIR=./cmd/arduino-mcp

ifeq ($(OS),Windows_NT)
	EXT=.exe
else
	EXT=
endif

.PHONY: all build clean test run

all: build

build:
	go build -o $(BUILD_DIR)/$(BINARY)$(EXT) $(CMD_DIR)

clean:
	rm -f $(BUILD_DIR)/$(BINARY)$(EXT)

test:
	go test ./internal/...

run: build
	./$(BINARY)$(EXT)

lint:
	go vet ./...
