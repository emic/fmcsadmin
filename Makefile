NAME=fmcsadmin
VERSION=1.0.1-dev

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
DIST_DIR=dist
LINUX_DIR=linux
MACOS_DIR=macos
WINDOWS_DIR=windows-x64
WINDOWS_32BIT_DIR=windows-x32
DIST_LINUX_DIR=$(NAME)-$(VERSION)-$(LINUX_DIR)
DIST_MACOS_DIR=$(NAME)-$(VERSION)-$(MACOS_DIR)
DIST_WINDOWS_DIR=$(NAME)-$(VERSION)-$(WINDOWS_DIR)
DIST_WINDOWS_32BIT_DIR=$(NAME)-$(VERSION)-$(WINDOWS_32BIT_DIR)

all: test build

deps:
	$(GOGET) github.com/mattn/go-scan
	$(GOGET) github.com/olekukonko/tablewriter
	$(GOGET) golang.org/x/crypto/ssh/terminal
	$(GOGET) github.com/stretchr/testify/assert

test: deps
	$(GOTEST) --cover

.PHONY: clean
clean:
	@rm -rf $(DIST_DIR)

build: build-linux build-macos build-windows build-windows-32bit

build-linux:
	mkdir -p $(DIST_DIR)/$(LINUX_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(LINUX_DIR)/$(NAME)

build-macos:
	mkdir -p $(DIST_DIR)/$(MACOS_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(MACOS_DIR)/$(NAME)

build-windows:
	mkdir -p $(DIST_DIR)/$(WINDOWS_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(WINDOWS_DIR)/$(NAME).exe

build-windows-32bit:
	mkdir -p $(DIST_DIR)/$(WINDOWS_32BIT_DIR)
	GOOS=windows GOARCH=386 CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(WINDOWS_32BIT_DIR)/$(NAME).exe

.PHONY: dist
dist: deps build
	cd $(DIST_DIR) && \
	mv $(LINUX_DIR) $(DIST_LINUX_DIR) && \
	cp -p ../LICENSE.txt $(DIST_LINUX_DIR)/ && \
	cp -p ../NOTICE.txt $(DIST_LINUX_DIR)/ && \
	cp -p ../README.md $(DIST_LINUX_DIR)/ && \
	cp -p ../release-notes.txt $(DIST_LINUX_DIR)/ && \
	tar -zcf $(DIST_LINUX_DIR).tar.gz $(DIST_LINUX_DIR) && \
	cd ..

	cd $(DIST_DIR) && \
	mv $(MACOS_DIR) $(DIST_MACOS_DIR) && \
	cp -p ../LICENSE.txt $(DIST_MACOS_DIR)/ && \
	cp -p ../NOTICE.txt $(DIST_MACOS_DIR)/ && \
	cp -p ../README.md $(DIST_MACOS_DIR)/ && \
	cp -p ../release-notes.txt $(DIST_MACOS_DIR)/ && \
	zip -r $(DIST_MACOS_DIR).zip $(DIST_MACOS_DIR) && \
	cd ..

	cd $(DIST_DIR) && \
	mv $(WINDOWS_DIR) $(DIST_WINDOWS_DIR) && \
	cp -p ../LICENSE.txt $(DIST_WINDOWS_DIR)/ && \
	cp -p ../NOTICE.txt $(DIST_WINDOWS_DIR)/ && \
	cp -p ../README.md $(DIST_WINDOWS_DIR)/ && \
	cp -p ../release-notes.txt $(DIST_WINDOWS_DIR)/ && \
	zip -r $(DIST_WINDOWS_DIR).zip $(DIST_WINDOWS_DIR) && \
	cd ..

	cd $(DIST_DIR) && \
	mv $(WINDOWS_32BIT_DIR) $(DIST_WINDOWS_32BIT_DIR) && \
	cp -p ../LICENSE.txt $(DIST_WINDOWS_32BIT_DIR)/ && \
	cp -p ../NOTICE.txt $(DIST_WINDOWS_32BIT_DIR)/ && \
	cp -p ../README.md $(DIST_WINDOWS_32BIT_DIR)/ && \
	cp -p ../release-notes.txt $(DIST_WINDOWS_32BIT_DIR)/ && \
	zip -r $(DIST_WINDOWS_32BIT_DIR).zip $(DIST_WINDOWS_32BIT_DIR) && \
	cd ..
