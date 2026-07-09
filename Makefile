BINARY_NAME=git-sage
INSTALL_DIR=/usr/local/bin
DIST_DIR=dist

.PHONY: all build dist clean install uninstall

all: build

build:
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(BINARY_NAME) main.go

dist:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go

install: build
	sudo cp $(DIST_DIR)/$(BINARY_NAME) $(INSTALL_DIR)

uninstall:
	sudo rm $(INSTALL_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(DIST_DIR)
