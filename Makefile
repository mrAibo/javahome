BINARY := javahome
CMD := ./cmd/javahome
BIN_DIR := bin
DIST_DIR := dist

.PHONY: help test build run install-local build-all clean

help:
	@echo "Available targets:"
	@echo "  make test          Run Go tests"
	@echo "  make build         Build $(BIN_DIR)/$(BINARY)"
	@echo "  make run           Run javahome from source"
	@echo "  make install-local Install into GOPATH/bin or GOBIN"
	@echo "  make build-all     Cross-compile common release binaries"
	@echo "  make clean         Remove build output"

test:
	go test ./...

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(CMD)

run:
	go run $(CMD) --help

install-local:
	go install $(CMD)

build-all:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY)-linux-amd64 $(CMD)
	GOOS=linux GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY)-linux-arm64 $(CMD)
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY)-darwin-amd64 $(CMD)
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY)-darwin-arm64 $(CMD)
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY)-windows-amd64.exe $(CMD)

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)
