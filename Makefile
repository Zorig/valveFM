APP_NAME := valvefm
BIN_DIR := bin

.DEFAULT_GOAL := help

.PHONY: help run build build-windows build-windows-gui tidy fmt clean

help:
	@echo "Targets:"
	@echo "  make run                Run TUI + tray"
	@echo "  make build              Build to bin/valvefm"
	@echo "  make build-windows      Build Windows console EXE"
	@echo "  make tidy               Run go mod tidy"
	@echo "  make fmt                Run gofmt"
	@echo "  make clean              Remove built binaries"

run:
	go run ./cmd/radio-tray

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/radio-tray

build-windows:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME).exe ./cmd/radio-tray

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal

clean:
	rm -f bin/$(APP_NAME) bin/$(APP_NAME).exe bin/$(APP_NAME)-gui.exe
