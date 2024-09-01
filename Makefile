VERSION=2.0.0
GOVERSION=$(shell go version)
USER=$(shell id -u -n)
TIME=$(shell date)
JR_HOME=jr

ifndef XDG_DATA_DIRS
ifeq ($(OS), Windows_NT)
	detectedOS := Windows
else
	detectedOS := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
endif

ifeq ($(detectedOS), Darwin)
	JR_SYSTEM_DIR="$(HOME)/Library/Application Support"
endif
ifeq ($(detectedOS),  Linux)
	JR_SYSTEM_DIR="$(HOME)/.config"
endif
ifeq ($(detectedOS), Windows_NT)
	JR_SYSTEM_DIR="$(LOCALAPPDATA)"
endif
else
	JR_SYSTEM_DIR=$(XDG_DATA_DIRS)
endif

ifndef XDG_DATA_HOME
ifeq ($(OS), Windows_NT)
	detectedOS := Windows
else
	detectedOS := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
endif

ifeq ($(detectedOS), Darwin)
	JR_USER_DIR="$(HOME)/.local/share"
endif
ifeq ($(detectedOS),  Linux)
	JR_USER_DIR="$(HOME)/.local/share"
endif
ifeq ($(detectedOS), Windows_NT)
	JR_USER_DIR="$(LOCALAPPDATA)" //@TODO
endif
else
	JR_USER_DIR=$(XDG_DATA_HOME)
endif

PLUGINS=mongodb \
        azblobstorage \
        azcosmosdb \
		luascript \
		awsdynamodb \
		cassandra

hello:
	@echo "JR Plugins"
	@echo " Version: $(VERSION)"
	@echo " Go Version: $(GOVERSION)"
	@echo " Build User: $(USER)"
	@echo " Build Time: $(TIME)"
	@echo " Detected OS: $(detectedOS)"
	@echo " JR System Dir: $(JR_SYSTEM_DIR)"
	@echo " JR User Dir: $(JR_USER_DIR)"


compile: hello lint test
	@echo "Compiling"
	for plugin in $(PLUGINS); do \
	    echo "building plugin jr-$$plugin"; \
		go build -v -ldflags="-X 'main.Version=$(VERSION)' \
		-X 'main.GoVersion=$(GOVERSION)' \
		-X 'main.BuildUser=$(USER)' \
		-X 'main.BuildTime=$(TIME)'" \
		-tags $$plugin \
		-o build/jr-$$plugin github.com/jrnd-io/jr-plugins/cmd/plugin; \
	done


clean:
	go clean
	rm build/*

test:
	go clean -testcache
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

dep:
	go mod download

vet:
	go vet

lint:
	golangci-lint run --config .localci/lint/golangci.yml

help: hello
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}all${RESET}'
	@echo ''


all: hello compile
all_offline: hello compile
