APP := media
BUILD_DIR := build
BIN := $(BUILD_DIR)/$(APP)

.PHONY: build run clean

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BIN) ./cmd/media

run:
	go run ./cmd/media

clean:
	rm -rf $(BUILD_DIR)
