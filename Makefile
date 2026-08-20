APP_NAME := valvefm
BIN_DIR := bin

.DEFAULT_GOAL := help

.PHONY: help run build build-tui build-windows build-windows-tui tidy fmt clean

help:
	@echo "Targets:"
	@echo "  make run                Run TUI + tray"
	@echo "  make build              Build to bin/valvefm (TUI + tray)"
	@echo "  make build-tui          Build TUI-only to bin/valvefm-tui"
	@echo "  make build-windows      Build Windows console EXE (TUI + tray)"
	@echo "  make build-windows-tui  Build Windows TUI-only EXE"
	@echo "  make tidy               Run go mod tidy"
	@echo "  make fmt                Run gofmt"
	@echo "  make clean              Remove built binaries"

run:
	go run ./cmd/radio-tray

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/radio-tray

build-tui:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME)-tui ./cmd/radio

build-windows:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME).exe ./cmd/radio-tray

build-windows-tui:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-tui.exe ./cmd/radio

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal

clean:
	rm -f bin/$(APP_NAME) bin/$(APP_NAME).exe bin/$(APP_NAME)-gui.exe
