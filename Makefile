NAME=fmcsadmin
VERSION=2.3.0

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
DIST_DIR=dist
LINUX_DIR=linux
LINUX_ARM64_DIR=linux-arm64
MACOS_DIR=macos
MACOS_ALT_DIR=macos-alt
WINDOWS_DIR=windows-x64
DIST_LINUX_DIR=$(NAME)-$(VERSION)-$(LINUX_DIR)
DIST_LINUX_ARM64_DIR=$(NAME)-$(VERSION)-$(LINUX_ARM64_DIR)
DIST_MACOS_DIR=$(NAME)-$(VERSION)-$(MACOS_DIR)
DIST_WINDOWS_DIR=$(NAME)-$(VERSION)-$(WINDOWS_DIR)

all: test build

deps:
	$(GOGET) github.com/golang-jwt/jwt/v5
	$(GOINSTALL) github.com/mattn/go-scan
	$(GOINSTALL) github.com/olekukonko/tablewriter
	$(GOINSTALL) golang.org/x/term
	$(GOINSTALL) github.com/stretchr/testify/assert

test: deps
	$(GOTEST) --cover

.PHONY: clean
clean:
	@rm -rf $(DIST_DIR)

build: build-linux build-linux-arm64 build-macos build-windows

build-linux:
	mkdir -p $(DIST_DIR)/$(LINUX_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(LINUX_DIR)/$(NAME)

build-linux-arm64:
	mkdir -p $(DIST_DIR)/$(LINUX_ARM64_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(LINUX_ARM64_DIR)/$(NAME)

ifeq ($(shell uname -m),x86_64)
build-macos:
	mkdir -p $(DIST_DIR)/$(MACOS_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(MACOS_DIR)/$(NAME)
else
build-macos:
	mkdir -p $(DIST_DIR)/$(MACOS_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(MACOS_DIR)/$(NAME)
endif

ifeq ($(shell uname -m),x86_64)
build-macos-alt:
	mkdir -p $(DIST_DIR)/$(MACOS_ALT_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(MACOS_ALT_DIR)/$(NAME)
else
build-macos-alt:
	mkdir -p $(DIST_DIR)/$(MACOS_ALT_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(MACOS_ALT_DIR)/$(NAME)
endif

build-windows:
	mkdir -p $(DIST_DIR)/$(WINDOWS_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(WINDOWS_DIR)/$(NAME).exe

.PHONY: dist
dist-multiplatform: deps build
	cd $(DIST_DIR) && \
	mv $(LINUX_DIR) $(DIST_LINUX_DIR) && \
	cp -p ../LICENSE.txt $(DIST_LINUX_DIR)/ && \
	cp -p ../NOTICE.txt $(DIST_LINUX_DIR)/ && \
	cp -p ../README.md $(DIST_LINUX_DIR)/ && \
	cp -p ../release-notes.txt $(DIST_LINUX_DIR)/ && \
	tar -zcf $(DIST_LINUX_DIR).tar.gz $(DIST_LINUX_DIR) && \
	cd ..

	cd $(DIST_DIR) && \
	mv $(LINUX_ARM64_DIR) $(DIST_LINUX_ARM64_DIR) && \
	cp -p ../LICENSE.txt $(DIST_LINUX_ARM64_DIR)/ && \
	cp -p ../NOTICE.txt $(DIST_LINUX_ARM64_DIR)/ && \
	cp -p ../README.md $(DIST_LINUX_ARM64_DIR)/ && \
	cp -p ../release-notes.txt $(DIST_LINUX_ARM64_DIR)/ && \
	tar -zcf $(DIST_LINUX_ARM64_DIR).tar.gz $(DIST_LINUX_ARM64_DIR) && \
	cd ..

	cd $(DIST_DIR) && \
	mv $(MACOS_DIR) $(DIST_MACOS_DIR) && \
	cp -p ../LICENSE.txt $(DIST_MACOS_DIR)/ && \
	cp -p ../NOTICE.txt $(DIST_MACOS_DIR)/ && \
	cp -p ../README.md $(DIST_MACOS_DIR)/ && \
	cp -p ../release-notes.txt $(DIST_MACOS_DIR)/ && \
	cd ..

	cd $(DIST_DIR) && \
	mv $(WINDOWS_DIR) $(DIST_WINDOWS_DIR) && \
	cp -p ../LICENSE.txt $(DIST_WINDOWS_DIR)/ && \
	cp -p ../NOTICE.txt $(DIST_WINDOWS_DIR)/ && \
	cp -p ../README.md $(DIST_WINDOWS_DIR)/ && \
	cp -p ../release-notes.txt $(DIST_WINDOWS_DIR)/ && \
	zip -r $(DIST_WINDOWS_DIR).zip $(DIST_WINDOWS_DIR) && \
	cd ..

ifeq ($(shell uname),Darwin)
dist: dist-multiplatform build-macos-alt
	cd $(DIST_DIR) && \
	mv $(DIST_MACOS_DIR)/$(NAME) $(DIST_MACOS_DIR)/$(NAME).tmp && \
	lipo -create $(DIST_MACOS_DIR)/$(NAME).tmp $(MACOS_ALT_DIR)/$(NAME) -output $(DIST_MACOS_DIR)/$(NAME) && \
	rm -f $(MACOS_ALT_DIR)/$(NAME) && \
	rmdir $(MACOS_ALT_DIR) && \
	rm -f $(DIST_MACOS_DIR)/$(NAME).tmp && \
	cd ..
else
dist: dist-multiplatform
endif