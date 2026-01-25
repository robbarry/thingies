.PHONY: build install clean test

BINARY_NAME=thingies
BUILD_DIR=bin
GO_FILES=$(shell find . -name '*.go' -type f)

build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/thingies

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)
	go clean

test:
	go test ./...

# Development helpers
run: build
	$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

fmt:
	go fmt ./...

tidy:
	go mod tidy
