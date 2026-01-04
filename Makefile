.PHONY: build install clean test

BINARY_NAME=tsp
INSTALL_DIR=$(HOME)/.local/bin

build:
	go build -o $(BINARY_NAME) ./cmd/tsp/

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)"

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test ./...

# Development helpers
run: build
	./$(BINARY_NAME)

fmt:
	go fmt ./...

lint:
	golangci-lint run

tidy:
	go mod tidy
