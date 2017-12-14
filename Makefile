GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=fmcsadmin
BINARY_DIR=bin
BINARY_LINUX_DIR=$(BINARY_DIR)/linux
BINARY_MACOS_DIR=$(BINARY_DIR)/macos
BINARY_WINDOWS_DIR=$(BINARY_DIR)/windows

all: test build

deps:
	$(GOGET) github.com/mattn/go-scan
	$(GOGET) github.com/olekukonko/tablewriter
	$(GOGET) golang.org/x/crypto/ssh/terminal

test: deps
	$(GOTEST) --cover

.PHONY: clean
clean:
	@rm -rf $(BINARY_DIR)

build: build-linux build-macos build-windows

build-linux:
	mkdir -p $(BINARY_LINUX_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_LINUX_DIR)/$(BINARY_NAME)

build-macos:
	mkdir -p $(BINARY_MACOS_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_MACOS_DIR)/$(BINARY_NAME)

build-windows:
	mkdir -p $(BINARY_WINDOWS_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_WINDOWS_DIR)/$(BINARY_NAME).exe
