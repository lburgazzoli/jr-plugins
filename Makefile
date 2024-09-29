VERSION=2.0.0
GOVERSION=$(shell go version)
USER=$(shell id -u -n)
TIME=$(shell date)
JR_HOME=jr

GOLANCI_LINT_VERSION ?= v1.61.0
GOVULNCHECK_VERSION ?= latest

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
LOCALBIN := $(PROJECT_PATH)/bin

ifndef XDG_DATA_DIRS
ifeq ($(OS), Windows_NT)
	detectedOS := Windows
else
	detectedOS := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
endif

ifeq ($(detectedOS), Darwin)
	JR_SYSTEM_DIR=/Library/Application Support
endif
ifeq ($(detectedOS),  Linux)
	JR_SYSTEM_DIR=/usr/local/share
endif
ifeq ($(detectedOS), Windows_NT)
	JR_SYSTEM_DIR=$(APPDATA)
endif
else
	JR_SYSTEM_DIR=$(XDG_DATA_DIRS[0])
endif

ifndef XDG_DATA_HOME
ifeq ($(OS), Windows_NT)
	detectedOS := Windows
else
	detectedOS := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
endif

ifeq ($(detectedOS), Darwin)
	JR_USER_DIR=$(HOME)/Library/Application Support
endif
ifeq ($(detectedOS),  Linux)
	JR_USER_DIR=$(HOME)/.local/share
endif
ifeq ($(detectedOS), Windows_NT)
	JR_USER_DIR=$(LOCALAPPDATA)
endif
else
	JR_USER_DIR=$(XDG_DATA_HOME)
endif

PLUGINS=mongodb \
        azblobstorage \
        azcosmosdb \
		luascript \
		awsdynamodb \
		s3 \
		cassandra \
		gcs \
		elastic \
		redis \
		http \
		wasm
comma:= ,
empty:=
space:= $(empty) $(empty)
prefix:= plugin_
build_tags:= $(subst $(space),$(comma),$(addprefix $(prefix),$(PLUGINS)))

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
		echo "building plugin $(prefix)$$plugin"; \
		go build -v -ldflags="-X 'main.Version=$(VERSION)' \
		-X 'main.GoVersion=$(GOVERSION)' \
		-X 'main.BuildUser=$(USER)' \
		-X 'main.BuildTime=$(TIME)'" \
		-tags $(prefix)$$plugin \
		-o build/jr-$$plugin github.com/jrnd-io/jr-plugins/cmd/plugin; \
	done


clean:
	go clean
	rm build/*

test:
	go clean -testcache
	go test -tags $(build_tags) ./...

test_coverage:
	go test  ./... -coverprofile=coverage.out

dep:
	go mod download

vet:
	go vet

.PHONY: lint
lint: golangci-lint
	$(LOCALBIN)/golangci-lint cache clean
	$(LOCALBIN)/golangci-lint run --config .localci/lint/golangci.yml

.PHONY: vuln
vuln: govulncheck
	$(LOCALBIN)/govulncheck -show verbose ./...

.PHONY: check
check: vet lint vuln

help: hello
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}all${RESET}'
	@echo ''

install: pluginsdir
	for plugin in $(PLUGINS); do \
	 echo "installing plugin jr-$$plugin"; \
	 install build/jr-$$plugin "$(JR_SYSTEM_DIR)/jr/plugins/"; \
	done

all: hello compile
all_offline: hello compile


$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: golangci-lint
golangci-lint: $(LOCALBIN)
	@test -s $(LOCALBIN)/golangci-lint || \
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANCI_LINT_VERSION)

.PHONY: govulncheck
govulncheck: $(LOCALBIN)
	@test -s $(LOCALBIN)/govulncheck || \
	GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

pluginsdir:
	mkdir -p "$(JR_SYSTEM_DIR)/jr/plugins"
