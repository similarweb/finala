# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTOOL=$(GOCMD) tool
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get


BINARY_NAME=finala
BINARY_LINUX=$(BINARY_NAME)_linux

DOCKER=docker
DOCKER_IMAGE=finala
DOCKER_TAG=dev


all: build

build: ## Download dependecies and Build the default binary 	 
		$(GOBUILD) -o $(BINARY_NAME) -v

test: ## Run tests for the project
		$(GOTEST) -coverprofile=cover.out -short -cover -failfast ./... | tee cov.txt
		
test-html: test ## Run tests with HTML for the project
		$(GOTOOL) cover -html=cover.out
		
build-linux: ## Build Cross Platform Binary
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_LINUX) -v

build-docker: ## BUild Docker image file
		$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

help: ## Show Help menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
