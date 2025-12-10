BINARY ?= git-opener
SRC_DIR := ./src

.PHONY: build run test clean

build:
	go build -o $(BINARY) $(SRC_DIR)

run: build
	./$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)
