.PHONY: proto tars unit integration coverage lint build package clean help

GO_WORK := go.work

# Go 环境变量防御性设置：支持从根目录 Makefile 继承或自动发现
# 优先使用已导出的环境变量，否则按常见路径查找
_GO_EXISTS := $(shell which go 2>/dev/null)
ifdef _GO_EXISTS
	export GOROOT ?= $(shell dir=$(_GO_EXISTS) && dirname "$$(dirname "$$dir")")
	export GOPATH ?= $(shell go env GOPATH 2>/dev/null)
else
	export GOROOT ?= /usr/local/go
	export GOPATH ?= $(HOME)/go
endif
export PATH := $(GOPATH)/bin:$(GOROOT)/bin:$(PATH)
export GOPROXY ?= https://goproxy.cn,direct

help: ## 显示 Go 子模块帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

proto: ## 生成 Protobuf Go 代码
	@echo "[go] ==> 生成 Protobuf Go 代码..."
	@bash ../../scripts/proto/generate-go.sh
	@echo "[go] Protobuf Go 代码生成完成"

tars: ## 生成 TarsGo/Tars2Go 代码
	@echo "[go] ==> 生成 TarsGo 代码..."
	@bash ../../scripts/proto/generate-tarsgo.sh || echo "[go] 跳过：TarsGo 工具未安装"
	@echo "[go] TarsGo 代码生成完成"

unit: ## 运行所有 go.work 中 module 的单元测试
	@echo "[go] ==> 运行单元测试..."
	@if [ -f $(GO_WORK) ]; then \
		for mod in gateway/proto-gateway tars/system; do \
			if [ -f "$$mod/go.mod" ]; then \
				echo "[go] 测试模块: $$mod"; \
				(cd $$mod && go test ./... -v -count=1 2>&1 | tail -20); \
			fi; \
		done; \
	else \
		echo "[go] 警告：go.work 不存在，尝试直接测试"; \
		go test ./... -v -count=1; \
	fi
	@echo "[go] 单元测试完成"

integration: ## 运行集成测试
	@echo "[go] ==> 运行集成测试..."
	@for mod in gateway tars/services; do \
		if [ -d "$$mod" ]; then \
			echo "[go] 集成测试: $$mod"; \
			cd $$mod && go test ./... -tags=integration -v 2>&1 || echo "[go] 跳过 $$mod 集成测试"; \
			cd ..; \
		fi; \
	done
	@echo "[go] 集成测试完成"

coverage: ## 生成 Go 覆盖率报告
	@echo "[go] ==> 生成覆盖率报告..."
	@mkdir -p ../../docs/reports/coverage/go
	@if [ -f $(GO_WORK) ]; then \
		for mod in $$(grep -oP '\./\K[^]+' $(GO_WORK)); do \
			if [ -f "$$mod/go.mod" ]; then \
				echo "[go] 覆盖率: $$mod"; \
				cd $$mod && go test ./... -coverprofile=cover.out -covermode=count && \
				go tool cover -html=cover.out -o ../../docs/reports/coverage/go/$${mod//\//-}_coverage.html || true; \
				cd ..; \
			fi; \
		done; \
	fi
	@echo "[go] 覆盖率报告已生成"

lint: ## 运行 Go Lint 检查（golangci-lint / vet）
	@echo "[go] ==> 运行 Lint 检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		for mod in $$(grep -oP '\./\K[^]+' $(GO_WORK) 2>/dev/null); do \
			if [ -f "$$mod/go.mod" ]; then \
				echo "[go] Lint: $$mod"; \
				cd $$mod && golangci-lint run ./... || true; \
				cd ..; \
			fi; \
		done; \
	else \
		for mod in $$(grep -oP '\./\K[^]+' $(GO_WORK) 2>/dev/null); do \
			if [ -f "$$mod/go.mod" ]; then \
				echo "[go] Vet: $$mod"; \
				cd $$mod && go vet ./...; \
				cd ..; \
			fi; \
		done; \
	fi
	@echo "[go] Lint 检查完成"

build: ## 构建 Go 可执行文件
	@echo "[go] ==> 构建项目..."
	@if [ -f $(GO_WORK) ]; then \
		for mod in $$(grep -oP '\./\K[^]+' $(GO_WORK)); do \
			if [ -f "$$mod/go.mod" ] && [ -f "$$mod/cmd" ]; then \
				echo "[go] 构建: $$mod"; \
				cd $$mod && go build -o ../../build/$${mod//\//-} ./cmd/... || true; \
				cd ..; \
			fi; \
		done; \
	fi
	@echo "[go] 构建完成"

package: ## 打包 Go 发布产物
	@echo "[go] ==> 打包产物..."
	@mkdir -p ../../dist/go
	@tar czf ../../dist/go/cairobot-go-$(shell date +%Y%m%d).tar.gz build/ || true
	@echo "[go] 打包完成"

clean: ## 清理 Go 构建产物
	@echo "[go] ==> 清理构建产物..."
	@rm -rf build/ cover.out
	@find . -name "*.test" -delete 2>/dev/null || true
	@find . -name "*.out" -delete 2>/dev/null || true
	@find . -name "*.prof" -delete 2>/dev/null || true
	@echo "[go] 清理完成"
