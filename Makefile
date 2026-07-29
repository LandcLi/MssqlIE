.PHONY: build test clean lint run help

APP_NAME     := mssql_ie
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT       ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE         ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS      := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
BUILD_DIR    := build

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## 编译项目
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .

build-all: ## 交叉编译 (linux + darwin)
	@mkdir -p $(BUILD_DIR)
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .

test: ## 运行测试
	go test -v -race -count=1 ./...

test-short: ## 运行简测 (无 race)
	go test -count=1 ./...

cover: ## 运行测试并生成覆盖率报告
	go test -v -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

clean: ## 清理构建产物
	rm -rf $(BUILD_DIR) coverage.out coverage.html *.out *.coverprofile

lint: ## 运行 lint 检查
	@golangci-lint run ./... 2>/dev/null || echo "提示: 请安装 golangci-lint (https://golangci-lint.run)"

fmt: ## 格式化代码
	go fmt ./...

vet: ## 运行 go vet
	go vet ./...

run: build ## 编译并运行（使用示例）
	@echo "运行: ./$(BUILD_DIR)/$(APP_NAME) --help"
	./$(BUILD_DIR)/$(APP_NAME) --help

check: lint vet test ## 运行完整检查 (lint + vet + test)
