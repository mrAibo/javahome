.PHONY: test build clean

test:
	go test ./...

build:
	go build -o bin/javahome ./cmd/javahome

clean:
	rm -rf bin dist
