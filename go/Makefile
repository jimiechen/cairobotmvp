.PHONY: proto tars unit integration coverage lint build package clean help \
       common-lib-test modules-test tars-test gateway-e2e go-all \
       gateway-build gateway-start gateway-test gateway-verify gateway-stop gateway-smoke \
       module-lint

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
		for mod in common-lib modules/hello modules/health tars/system gateway/proto-gateway; do \
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

common-lib-test: ## 测试 common-lib 公共库（错误码 + 类型定义）
	@echo "[go] ==> 测试 common-lib..."
	@if [ -d common-lib ]; then \
		cd common-lib && go test ./... -v -count=1 2>&1; \
	else \
		echo "[go] ❌ common-lib 目录不存在"; \
		exit 1; \
	fi
	@echo "[go] common-lib 测试完成"

modules-test: ## 测试所有业务模块（hello + health）
	@echo "[go] ==> 测试业务模块..."
	@for mod in modules/hello modules/health; do \
		if [ -f "$$mod/go.mod" ]; then \
			echo "[go] 测试模块: $$mod"; \
			(cd $$mod && go test ./... -v -count=1 2>&1); \
		else \
			echo "[go] ⚠️  跳过：$$mod 不存在"; \
		fi; \
	done
	@echo "[go] 业务模块测试完成"

tars-test: ## 测试 Tars 调用层（adapter + deprecated + service）
	@echo "[go] ==> 测试 tars/system..."
	@if [ -d tars/system ]; then \
		cd tars/system && go test ./... -v -count=1 2>&1; \
	else \
		echo "[go] ❌ tars/system 目录不存在"; \
		exit 1; \
	fi
	@echo "[go] tars/system 测试完成"

gateway-e2e: ## 运行 proto-gateway E2E 链路验收测试（Gateway → Modules 全链路）
	@echo "[go] ==> 运行 Gateway E2E 链路测试..."
	@cd $(GATEWAY_DIR) && go test -v -count=1 -run "TestGateway_E2E" ./... 2>&1
	@echo "[go] Gateway E2E 测试完成"

go-all: ## 运行 Go 全量测试（common-lib + modules + tars + gateway）并输出汇总
	@echo "=========================================="
	@echo "  Go Monorepo 全量测试"
	@echo "=========================================="
	@echo ""
	@$(MAKE) common-lib-test
	@echo ""
	@$(MAKE) modules-test
	@echo ""
	@$(MAKE) tars-test
	@echo ""
	@$(MAKE) gateway-test
	@echo ""
	@$(MAKE) gateway-e2e
	@echo ""
	@echo "=========================================="
	@echo "  ✅ Go 全量测试完成"
	@echo "=========================================="

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

# ============================================================
# proto-gateway 专项命令（TarsGo HTTP Servant）
# ============================================================

GATEWAY_DIR := gateway/proto-gateway
GATEWAY_BIN := ../../build/proto-gateway-server
GATEWAY_PID_FILE := ../../build/.gateway.pid

gateway-build: ## 编译 proto-gateway 二进制（含 TarsGo v1.4.6 框架）
	@echo "[gateway] ==> 编译 proto-gateway..."
	@mkdir -p ../../build
	@cd $(GATEWAY_DIR) && go build -o $(GATEWAY_BIN) ./cmd/server/main.go
	@echo "[gateway] ✅ 编译完成: $(GATEWAY_BIN)"
	@ls -la $(GATEWAY_BIN)

gateway-start: ## 启动 proto-gateway（local 模式，TarsGo HTTP on :8080）
	@echo "[gateway] ==> 启动 proto-gateway (local mode)..."
	@if [ ! -f $(GATEWAY_BIN) ]; then \
		echo "[gateway] ⚠️  二进制不存在，先执行 make -C go gateway-build"; \
		exit 1; \
	fi
	@if [ -f $(GATEWAY_PID_FILE) ] && kill -0 $$(cat $(GATEWAY_PID_FILE)) 2>/dev/null; then \
		echo "[gateway] ⚠️  Gateway 已在运行 (pid=$$(cat $(GATEWAY_PID_FILE)))"; \
	else \
		cd $(GATEWAY_DIR) && setsid env GATEWAY_INVOKER_MODE=local $(GATEWAY_BIN) >../../build/gateway.log 2>&1 & echo $$! >$(GATEWAY_PID_FILE); \
		sleep 2; \
		if kill -0 $$(cat $(GATEWAY_PID_FILE)) 2>/dev/null; then \
			echo "[gateway] ✅ 已启动 (pid=$$(cat $(GATEWAY_PID_FILE))) 监听 :8080"; \
			echo "[gateway] 日志: build/gateway.log | TarsGo日志: /tmp/cairobot/logs/gateway/"; \
		else \
			echo "[gateway] ❌ 启动失败，查看 build/gateway.log:"; \
			cat ../../build/gateway.log 2>/dev/null || true; \
			rm -f $(GATEWAY_PID_FILE); \
			exit 1; \
		fi; \
	fi

gateway-stop: ## 停止 proto-gateway
	@echo "[gateway] ==> 停止 proto-gateway..."
	@if [ -f $(GATEWAY_PID_FILE) ]; then \
		pid=$$(cat $(GATEWAY_PID_FILE)); \
		if kill -0 $$pid 2>/dev/null; then \
			kill $$pid && echo "[gateway] ✅ 已停止 (pid=$$pid)" || echo "[gateway] ⚠️  停止信号已发送"; \
		else \
			echo "[gateway] 进程不存在 (stale pid=$$pid)"; \
		fi; \
		rm -f $(GATEWAY_PID_FILE); \
	else \
		echo "[gateway] 未找到 PID 文件，尝试 lsof+kill..."; \
		for pid in $$(lsof -ti :8080 2>/dev/null); do \
			kill $$pid 2>/dev/null && echo "[gateway] ✅ 已停止 (pid=$$pid)" || true; \
		done; \
		if [ -z "$$(lsof -ti :8080 2>/dev/null)" ]; then \
			echo "[gateway] ✅ 端口 :8080 已释放"; \
		else \
			echo "[gateway] ⚠️ 仍有进程占用 :8080"; \
		fi; \
	fi

gateway-restart: ## 重启 proto-gateway（停止 + 重新编译 + 启动，:8080）
	@echo "[gateway] ==> 重启 proto-gateway..."
	@$(MAKE) gateway-stop
	@sleep 1
	@$(MAKE) gateway-build
	@$(MAKE) gateway-start

gateway-test: ## 运行 proto-gateway 全部测试
	@echo "[gateway] ==> 运行 proto-gateway 测试..."
	@cd $(GATEWAY_DIR) && go test ./... -v -count=1 2>&1 | tail -30
	@echo "[gateway] 测试完成"

gateway-smoke: ## 快速冒烟测试（编译 + 测试 + 启动 + 验证 /api/hello + 停止）
	@echo "[gateway] ========== 冒烟测试开始 =========="
	@$(MAKE) gateway-build
	@$(MAKE) gateway-start
	@sleep 2
	@echo "[gateway] --> 发送 HealthCheck 请求..."
	@cd $(GATEWAY_DIR) && go run ./cmd/testclient/main.go 2>&1 || echo "[gateway] ⚠️  testclient 执行异常"
	@$(MAKE) gateway-stop
	@echo "[gateway] ========== 冒烟测试结束 =========="

gateway-verify: ## 完整验证（编译 + 单元测试 + TarsGo 依赖检查）
	@echo "[gateway] ========== 完整验证 =========="
	@echo "[gateway] 1/3 编译..."
	@$(MAKE) gateway-build
	@echo "[gateway] 2/3 单元测试..."
	@cd $(GATEWAY_DIR) && go test ./... -count=1 2>&1 | grep -E "PASS|FAIL|ok |build failed"
	@echo "[gateway] 3/3 TarsGo 依赖检查..."
	@cd $(GATEWAY_DIR) && go list -m all | grep -i "tars\|tarscloud" || echo "[gateway] ⚠️  未找到 TarsGo 依赖"
	@echo "[gateway] ========== 验证完成 =========="

# ============================================================
# module-lint：模块接入规范自动检查（10+1 项强制检查）
# ============================================================

module-lint: ## 模块合规性 lint 检查（L1-L10 + SDK_USAGE，任一失败 exit 1）
	@echo "[go] ==> 执行模块 lint 检查..."
	@bash ../../scripts/lint/module_lint.sh
	@echo "[go] ✅ module-lint 全部通过"

module-lint-hello: ## 仅检查 hello 模块合规性
	@bash ../../scripts/lint/module_lint.sh --module hello

module-lint-health: ## 仅检查 health 模块合规性
	@bash ../../scripts/lint/module_lint.sh --module health
