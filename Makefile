# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTOOL=$(GOCMD) tool
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD)fmt

GOARCH=$(shell go env GOARCH)

BINARY_NAME=finala
BINARY_LINUX=$(BINARY_NAME)_linux

DOCKER=docker
DOCKER_IMAGE=finala
DOCKER_TAG=dev

TEST_EXEC_CMD=$(GOTEST) -coverprofile=cover.out -short -cover -failfast ./... 

all: build

build: build-ui ## Download dependecies and Build the default binary
		$(GOBUILD) -o $(BINARY_NAME) -v

build-ui: ## Build applicaiton UI
		cd ui && npm install && npm run build && cd ..

test: ## Run tests for the project
		$(TEST_EXEC_CMD)
		
test-html:  ## Run tests with HTML for the project
		$(TEST_EXEC_CMD) | true
		$(GOTOOL) cover -html=cover.out

checks: test lint fmt ## Run all tests,linting and FMT

lint:
	golangci-lint run

fmt: ## Validate go format
	@echo checking gofmt...
	@res=$$($(GOFMT) -d -e -s $$(find . -type d \( -path ./src/vendor \) -prune -o -name '*.go' -print)); \
	if [ -n "$${res}" ]; then \
		echo checking gofmt fail... ; \
		echo "$${res}"; \
		exit 1; \
	else \
		echo Your code formating is according gofmt standards; \
	fi
		
build-linux: ## Build Cross Platform Binary
		CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) $(GOBUILD) -o $(BINARY_NAME)_linux -v

build-osx: ## Build Mac Binary
		CGO_ENABLED=0 GOOS=darwin GOARCH=$(GOARCH) $(GOBUILD) -o $(BINARY_NAME)_osx -v

build-windows: ## Build Windows Binary
		CGO_ENABLED=0 GOOS=windows GOARCH=$(GOARCH) $(GOBUILD) -o $(BINARY_NAME)_windows -v

build-docker: ## BUild Docker image file
		$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

help: ## Show Help menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
