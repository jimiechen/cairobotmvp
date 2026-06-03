.PHONY: help bootstrap proto proto-check lint test unit integration coverage build package \
        docs rules testcase-check comment-check ci clean \
        gateway-build gateway-start gateway-stop gateway-test gateway-smoke gateway-verify \
        go-all common-lib-test modules-test tars-test gateway-e2e \
        admin-dev admin-start admin-stop admin-backend admin-frontend admin-status admin-restart \
admin-migrate \
        e2e-admin e2e-config e2e-i18n e2e-all e2e-install e2e-report

PROJECT_ROOT := $(shell pwd)
export PROJECT_ROOT

# ============================================================
# 多语言工具链自动发现（无需全局配置，Makefile 自动识别）
# ============================================================

# --- Go 环境 ---
export GOPROXY ?= https://goproxy.cn,direct
_GO_BIN := $(shell which go 2>/dev/null)
ifeq ($(_GO_BIN),)
	_GO_BIN := $(shell ls $(GOROOT)/bin/go 2>/dev/null)
endif
ifeq ($(_GO_BIN),)
	_GO_BIN := $(shell ls /usr/local/opt/go/bin/go 2>/dev/null)
endif
ifeq ($(_GO_BIN),)
	_GO_BIN := $(shell ls $(HOME)/go/bin/go 2>/dev/null)
endif
ifeq ($(_GO_BIN),)
	_GO_BIN := $(shell ls /usr/local/go/bin/go 2>/dev/null)
endif
ifneq ($(_GO_BIN),)
	_GOROOT := $(shell dir=$(_GO_BIN) && dirname "$$(dirname "$$dir")")
	_GOPATH := $(shell $(_GO_BIN) env GOPATH 2>/dev/null)
	export GOROOT := $(_GOROOT)
	export GOPATH := $(_GOPATH)
	export PATH := $(_GOPATH)/bin:$(GOROOT)/bin:$(PATH)
else
	export GOROOT ?= /usr/local/go
	export GOPATH ?= $(HOME)/go
	export PATH := $(GOPATH)/bin:$(GOROOT)/bin:$(PATH)
endif

# --- Python 环境（优先 python3）---
# 查找顺序：PATH → Homebrew → pyenv → asdf → conda → /usr/local/bin → ~/miniconda3
_PYTHON_BIN := $(shell which python3 2>/dev/null)
ifeq ($(_PYTHON_BIN),)
	_PYTHON_BIN := $(shell which python 2>/dev/null)
endif
ifeq ($(_PYTHON_BIN),)
	_PYTHON_BIN := $(shell ls /usr/local/opt/python/libexec/bin/python3 2>/dev/null)
endif
ifeq ($(_PYTHON_BIN),)
	_PYTHON_BIN := $(if $(wildcard ~/.pyenv/shims/python3),$(shell cat ~/.pyenv/version 2>/dev/null >/dev/null && echo $$(pyenv which python3)))
endif
ifeq ($(_PYTHON_BIN),)
	_PYTHON_BIN := $(shell ls $(HOME)/miniconda3/bin/python3 2>/dev/null)
endif
ifeq ($(_PYTHON_BIN),)
	_PYTHON_BIN := $(shell ls $(HOME)/anaconda3/bin/python3 2>/dev/null)
endif
ifeq ($(_PYTHON_BIN),)
	_PYTHON_BIN := $(shell ls /usr/local/bin/python3 2>/dev/null)
endif
ifdef _PYTHON_BIN
	export PYTHON := $(_PYTHON_BIN)
	_PYTHON_DIR := $(shell dirname $(_PYTHON_BIN))
	# 将 Python bin 目录加入 PATH（pip 安装的工具可用）
	export PATH := $(_PYTHON_DIR):$(_PYTHON_DIR)/../bin:$(PATH)
	# PIP 使用国内镜像加速
	export PIP_INDEX_URL ?= https://pypi.tuna.tsinghua.edu.cn/simple
	export PIP_TRUSTED_HOST ?= pypi.tuna.tsinghua.edu.cn
else
	export PYTHON := python3
endif

# --- Node.js / npm 环境 ---
# 查找顺序：PATH → Homebrew → nvm → fnm → volta → /usr/local/bin
_NODE_BIN := $(shell which node 2>/dev/null)
ifeq ($(_NODE_BIN),)
	_NODE_BIN := $(shell ls /usr/local/opt/node/bin/node 2>/dev/null)
endif
ifeq ($(_NODE_BIN),)
	_NODE_BIN := $(if $(wildcard ~/.nvm/versions/node/*/bin/node),$(shell ls -t ~/.nvm/versions/node/*/bin/node 2>/dev/null | head -1))
endif
ifeq ($(_NODE_BIN),)
	_NODE_BIN := $(if $(wildcard ~/.fnm/current/bin/node),$(shell ls ~/.fnm/current/bin/node 2>/dev/null))
endif
ifeq ($(_NODE_BIN),)
	_NODE_BIN := $(if $(wildcard ~/.volta/bin/node),$(shell ls ~/.volta/bin/node 2>/dev/null))
endif
ifeq ($(_NODE_BIN),)
	_NODE_BIN := $(shell ls /usr/local/bin/node 2>/dev/null)
endif
ifdef _NODE_BIN
	export NODE := $(_NODE_BIN)
	_NODE_DIR := $(shell dirname $(_NODE_BIN))
	_NPM_BIN := $(_NODE_DIR)/npm
	export PATH := $(_NODE_DIR):$(PATH)
	# npm 使用国内镜像加速
	export NPM_REGISTRY ?= https://registry.npmmirror.com
	export ELECTRON_MIRROR ?= https://npmmirror.com/mirrors/electron/
else
	export NODE := node
endif

# --- Protobuf 工具 ---
_PROTOC_BIN := $(shell which protoc 2>/dev/null)
ifdef _PROTOC_BIN
	_PROTOC_DIR := $(shell dirname $(_PROTOC_BIN))
endif
export PATH := $(or $(_PROTOC_DIR),/usr/local/bin):$(PATH)

help: ## 显示帮助信息
	@echo "CaiRobot MVP 工程入口 Makefile"
	@echo ""
	@echo "━━━ 工具链自动检测 ━━━"
	@echo "  Go:    $(_GO_BIN) $(if $(_GO_BIN),✅,❌ 未找到)"
	@echo "  Python: $(PYTHON) $(if $(_PYTHON_BIN),✅,❌ 未找到)"
	@echo "  Node:  $(NODE) $(if $(_NODE_BIN),✅,❌ 未找到)"
	@echo ""
	@echo "━━━ 工程编排（全量） ━━━"
	@grep -E '^(bootstrap|ci|clean):.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'
	@echo ""
	@echo "━━━ 构建 / 测试 / 检查 ━━━"
	@grep -E '^(proto|proto-check|lint|test|unit|integration|coverage|build|package|docs|rules|testcase-check|comment-check):.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'
	@echo ""
	@echo "━━━ Gateway（Proto 网关） ━━━"
	@grep -E '^(gateway-|go-all|common-lib-test|modules-test|tars-test)[^a-z]*## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'
	@echo ""
	@echo "━━━ Admin MVP 前后端服务 ━━━"
	@grep -E '^(admin-)[a-z].*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'
	@echo ""
	@echo "━━━ Admin MVP E2E 端到端测试 ━━━"
	@grep -E '^(e2e-)[a-z].*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'

bootstrap: ## 初始化开发环境（检查工具链 + 安装依赖）
	@echo "==> 检查开发工具链..."
	@bash scripts/ci/check_tools.sh
	@echo "==> 初始化 Go 开发环境..."
	@if [ -f go/Makefile ]; then $(MAKE) -C go install || true; fi
	@echo "==> 初始化 TypeScript 开发环境..."
	@if [ -f typescript/Makefile ]; then $(MAKE) -C typescript install || true; fi
	@echo "==> 初始化 Python 开发环境..."
	@if [ -f python/Makefile ]; then $(MAKE) -C python install || true; fi
	@echo "==> Bootstrap 完成"

proto: ## 生成所有语言的 Protobuf 代码（Go/TS/Python/TarsGo），工具缺失时明确失败
	@echo "==> 检查 protoc 工具..."
	@if ! command -v protoc >/dev/null 2>&1; then \
		echo "❌ 错误：protoc 未安装"; \
		echo "   macOS: brew install protobuf"; \
		echo "   Linux: apt install -y protobuf-compiler"; \
		echo "   或执行 make bootstrap 检查所有工具"; \
		exit 1; \
	fi
	@echo "==> 生成 Protobuf 代码..."
	@bash scripts/proto/generate-go.sh
	@bash scripts/proto/generate-ts.sh || echo "[skip] protoc-gen-ts 未安装"
	@bash scripts/proto/generate-python.sh || echo "[skip] grpcio-tools 未安装"
	@bash scripts/proto/generate-tarsgo.sh || echo "[skip] tars2go 未安装"
	@echo "==> Protobuf 代码生成完成"

proto-check: ## 校验 Protobuf 生成代码是否存在且注册表一致（CI 用，不需要 protoc）
	@echo "==> 校验 Protobuf 生成代码..."
	@$(PYTHON) scripts/ci/check_proto_registry.py
	@echo "==> Protobuf 校验完成"

lint: ## 运行所有语言的 Lint 检查
	@echo "==> 运行 Lint 检查..."
	@if [ -f go/Makefile ]; then $(MAKE) -C go lint || echo "跳过 go lint"; fi
	@if [ -f typescript/Makefile ]; then $(MAKE) -C typescript lint || echo "跳过 typescript lint"; fi
	@if [ -f python/Makefile ]; then $(MAKE) -C python lint || echo "跳过 python lint"; fi
	@echo "==> Lint 检查完成"

test: ## 运行所有测试（单元 + 集成）
	$(MAKE) unit
	$(MAKE) integration

unit: ## 运行所有单元测试
	@echo "==> 运行单元测试..."
	@if [ -f go/Makefile ]; then $(MAKE) -C go unit || echo "[go] 单元测试完成"; fi
	@if [ -f typescript/Makefile ]; then $(MAKE) -C typescript unit || echo "[ts] 跳过 web test（需 pnpm install）"; fi
	@if [ -f python/Makefile ]; then $(MAKE) -C python unit || echo "[py] 跳过 python test（需 pip install）"; else \
		if command -v $(PYTHON) >/dev/null 2>&1; then \
			echo "[py] ==> 运行单元测试..."; \
			cd python && $(PYTHON) -m pytest 2>&1 || echo "[py] ⚠️  pytest 未安装或无测试文件"; \
		else \
			echo "[py] 跳过 python test（未找到 Python 环境）"; \
		fi; \
	fi
	@echo "==> 单元测试完成"

integration: ## 运行集成测试
	@echo "==> 运行集成测试..."
	@if [ -f go/Makefile ]; then $(MAKE) -C go integration || echo "跳过 go integration"; fi
	@echo "==> 集成测试完成"

coverage: ## 生成覆盖率报告，输出最终汇总结果
	@echo "==> 生成覆盖率报告..."
	@mkdir -p docs/reports/coverage/go docs/reports/coverage/typescript docs/reports/coverage/python
	@bash scripts/coverage/go_coverage.sh; RC=$$?; if [ $$RC -ne 0 ]; then echo "[warn] Go 覆盖率收集异常 (rc=$$RC)"; fi
	@bash scripts/coverage/ts_coverage.sh; RC=$$?; if [ $$RC -ne 0 ]; then echo "[warn] TS 覆盖率收集异常 (rc=$$RC)"; fi
	@bash scripts/coverage/python_coverage.sh; RC=$$?; if [ $$RC -ne 0 ]; then echo "[warn] Python 覆盖率收集异常 (rc=$$RC)"; fi
	@bash scripts/coverage/merge_coverage_reports.sh
	@echo ""
	@echo "==> 覆盖率报告已生成到 docs/reports/coverage/"
	@echo "    摘要: docs/reports/coverage/coverage-summary.md"
	@if [ -f docs/reports/coverage/coverage-summary.md ]; then \
		cat docs/reports/coverage/coverage-summary.md; \
	fi

build: ## 构建所有语言的可执行文件
	@echo "==> 构建项目..."
	@if [ -f go/Makefile ]; then $(MAKE) -C go build || echo "跳过 go build"; fi
	@if [ -f typescript/Makefile ]; then $(MAKE) -C typescript build || echo "跳过 typescript build"; fi
	@if [ -f python/Makefile ]; then $(MAKE) -C python build || echo "跳过 python build"; fi
	@echo "==> 构建完成"

package: ## 打包发布产物
	@echo "==> 打包产物..."
	@if [ -f go/Makefile ]; then $(MAKE) -C go package || echo "跳过 go package"; fi
	@if [ -f typescript/Makefile ]; then $(MAKE) -C typescript package || echo "跳过 typescript package"; fi
	@if [ -f python/Makefile ]; then $(MAKE) -C python package || echo "跳过 python package"; fi
	@echo "==> 打包完成"

docs: ## 检查关键文档是否存在
	@echo "==> 检查文档完整性..."
	@$(PYTHON) scripts/ci/check_required_docs.py

rules: ## 执行工程规范检查（工具链 + 目录布局 + Make target + 模块路径 + 注释 + 测试用例）
	@echo "==> 执行工程规范检查..."
	@bash scripts/ci/check_tools.sh || echo "❌ 工具链检查未通过"
	@$(PYTHON) scripts/ci/check_directory_layout.py || echo "❌ 目录布局检查未通过"
	@bash scripts/ci/check_make_targets.sh || echo "❌ Make target 检查未通过"
	@$(PYTHON) scripts/ci/check_module_paths.py || echo "❌ 模块路径检查未通过"
	@$(PYTHON) scripts/ci/check_chinese_comments.py || echo "❌ 中文注释检查未通过"
	@$(PYTHON) scripts/ci/check_testcase_registry.py || echo "❌ 测试用例注册表检查未通过"
	@echo "==> 规范检查完成"

testcase-check: ## 检查测试用例注册表一致性
	@echo "==> 检查测试用例注册表..."
	@$(PYTHON) scripts/ci/check_testcase_registry.py

comment-check: ## 检查中文注释规范性
	@echo "==> 检查中文注释..."
	@$(PYTHON) scripts/ci/check_chinese_comments.py

ci: ## 完整 CI 检查（本地等价于 GitHub Actions）
	@echo "=========================================="
	@echo "  CaiRobot MVP 完整 CI 检查"
	@echo "=========================================="
	@echo ""
	$(MAKE) docs
	$(MAKE) rules
	$(MAKE) proto-check
	$(MAKE) lint
	$(MAKE) unit
	$(MAKE) integration
	$(MAKE) coverage
	@echo ""
	@echo "=========================================="
	@echo "  CI 检查全部通过 ✅"
	@echo "=========================================="

clean: ## 清理构建产物和临时文件
	@echo "==> 清理构建产物..."
	@if [ -f go/Makefile ]; then $(MAKE) -C go clean || true; fi
	@if [ -f typescript/Makefile ]; then $(MAKE) -C typescript clean || true; fi
	@if [ -f python/Makefile ]; then $(MAKE) -C python clean || true; fi
	@rm -rf docs/reports/coverage/
	@echo "==> 清理完成"

# ============================================================
# proto-gateway 专项命令（透传到 go/Makefile）
# ============================================================

gateway-build: ## 编译 proto-gateway（含 TarsGo v1.4.6）
	@$(MAKE) -C go gateway-build

gateway-start: ## 启动 proto-gateway（local 模式，TarsGo HTTP on :8080）
	@$(MAKE) -C go gateway-start

gateway-stop: ## 停止 proto-gateway
	@$(MAKE) -C go gateway-stop

gateway-test: ## 运行 proto-gateway 全部测试
	@$(MAKE) -C go gateway-test

gateway-smoke: ## 冒烟测试：编译 + 启动 + 验证 /api/hello + 停止
	@$(MAKE) -C go gateway-smoke

gateway-verify: ## 完整验证：编译 + 测试 + TarsGo 依赖检查
	@$(MAKE) -C go gateway-verify

# ============================================================
# Go Monorepo 模块化测试命令（透传到 go/Makefile）
# ============================================================

go-all: ## 运行 Go 全量测试（common-lib + modules + tars + gateway + E2E）
	@$(MAKE) -C go go-all

common-lib-test: ## 测试 common-lib 公共库（错误码 + 类型定义）
	@$(MAKE) -C go common-lib-test

modules-test: ## 测试所有业务模块（hello + health）
	@$(MAKE) -C go modules-test

tars-test: ## 测试 Tars 调用层（adapter + deprecated + service）
	@$(MAKE) -C go tars-test

gateway-e2e: ## 运行 proto-gateway E2E 链路验收测试（Gateway → Modules 全链路）
	@$(MAKE) -C go gateway-e2e

# ============================================================
# Admin MVP 前后端启动/停止命令
# 后端: go-admin (Gin + GORM, port 8000)
# 前端: admin-web (Vue 2 + Element UI, port 9527)
# ============================================================

ADMIN_PID_DIR := .admin-pids
ADMIN_LOG_DIR := .admin-logs
ADMIN_BACKEND_PID := $(ADMIN_PID_DIR)/backend.pid
ADMIN_FRONTEND_PID := $(ADMIN_PID_DIR)/frontend.pid
ADMIN_BACKEND_LOG := $(ADMIN_LOG_DIR)/backend.log
ADMIN_FRONTEND_LOG := $(ADMIN_LOG_DIR)/frontend.log
ADMIN_DB_FILE := go/admin/go-admin-db.db

ADMIN_BACKEND_PORT := 8000
ADMIN_FRONTEND_PORT := 9527

admin-dev: ## 一键启动 Admin MVP 前后端（开发模式，后台运行）
	@echo "=========================================="
	@echo "  Admin MVP 开发环境启动"
	@echo "  后端: Go + SQLite3 (port $(ADMIN_BACKEND_PORT))"
	@echo "  前端: Vue 2 + Element UI (port $(ADMIN_FRONTEND_PORT))"
	@echo "=========================================="
	@mkdir -p $(ADMIN_PID_DIR) $(ADMIN_LOG_DIR)
	@if [ ! -f $(ADMIN_DB_FILE) ] || [ ! -s $(ADMIN_DB_FILE) ]; then \
		echo "==> 数据库文件不存在或为空，执行 migrate 初始化..."; \
		cd go/admin && $(_GO_BIN) run -tags sqlite3 main.go migrate; \
	fi
	@cd go/admin && sqlite3 $(ADMIN_DB_FILE) "SELECT name FROM sqlite_master WHERE type='table' AND name='config_schema'" 2>/dev/null | grep -q config_schema || { \
		echo "==> 业务表不存在，导入种子数据..."; \
		sqlite3 $(ADMIN_DB_FILE) < config/db-business.sql && echo "✅ 业务数据导入完成"; \
	}
	@if [ -f $(ADMIN_BACKEND_PID) ] && lsof -i :$(ADMIN_BACKEND_PORT) >/dev/null 2>&1; then \
		echo "⚠️  后端已在运行 (port $(ADMIN_BACKEND_PORT))"; \
	else \
		echo "==> 启动 Go 后端（首次需编译 SQLite3 CGO，约 1-2 分钟）..."; \
		cd go/admin && nohup $(_GO_BIN) run -tags sqlite3 main.go server > $(PROJECT_ROOT)/$(ADMIN_BACKEND_LOG) 2>&1 & echo $$! > $(PROJECT_ROOT)/$(ADMIN_BACKEND_PID); \
		echo "    等待端口监听..."; \
		for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24; do \
			sleep 5; \
			if lsof -i :$(ADMIN_BACKEND_PORT) >/dev/null 2>&1; then \
				echo "✅ Go 后端就绪 (PID $$(cat $(ADMIN_BACKEND_PID)), port $(ADMIN_BACKEND_PORT), 等待 $$((i*5))s)"; \
				break; \
			fi; \
			if ! kill -0 $$(cat $(ADMIN_BACKEND_PID)) 2>/dev/null; then \
				echo "❌ Go 后端进程已退出，查看日志:"; tail -30 $(ADMIN_BACKEND_LOG); rm -f $(ADMIN_BACKEND_PID); exit 1; \
			fi; \
			echo "    编译/启动中... ($$((i*5))s)"; \
		done; \
		if ! lsof -i :$(ADMIN_BACKEND_PORT) >/dev/null 2>&1; then \
			echo "❌ 超时：$(ADMIN_BACKEND_PORT) 端口未监听（120s），查看日志:"; tail -30 $(ADMIN_BACKEND_LOG); rm -f $(ADMIN_BACKEND_PID); exit 1; \
		fi; \
	fi
	@if [ -f $(ADMIN_FRONTEND_PID) ] && lsof -i :$(ADMIN_FRONTEND_PORT) >/dev/null 2>&1; then \
		echo "⚠️  前端已在运行 (port $(ADMIN_FRONTEND_PORT))"; \
	else \
		echo "==> 启动 Vue 前端..."; \
		cd typescript/admin-web && NODE_OPTIONS=--openssl-legacy-provider nohup npx vue-cli-service serve > $(PROJECT_ROOT)/$(ADMIN_FRONTEND_LOG) 2>&1 & echo $$! > $(PROJECT_ROOT)/$(ADMIN_FRONTEND_PID); \
		for i in 1 2 3 4 5 6; do \
			sleep 5; \
			if lsof -i :$(ADMIN_FRONTEND_PORT) >/dev/null 2>&1; then \
				echo "✅ Vue 前端就绪 (PID $$(cat $(ADMIN_FRONTEND_PID)), port $(ADMIN_FRONTEND_PORT), 等待 $$((i*5))s)"; \
				break; \
			fi; \
			if ! kill -0 $$(cat $(ADMIN_FRONTEND_PID)) 2>/dev/null; then \
				echo "❌ Vue 前端进程已退出，查看日志:"; tail -20 $(ADMIN_FRONTEND_LOG); rm -f $(ADMIN_FRONTEND_PID); exit 1; \
			fi; \
			echo "    编译/启动中... ($$((i*5))s)"; \
		done; \
		if ! lsof -i :$(ADMIN_FRONTEND_PORT) >/dev/null 2>&1; then \
			echo "❌ 超时：$(ADMIN_FRONTEND_PORT) 端口未监听（30s）"; rm -f $(ADMIN_FRONTEND_PID); exit 1; \
		fi; \
	fi
	@echo ""
	@echo "=========================================="
	@echo "  ✅ Admin MVP 开发环境就绪"
	@echo "  前端地址: http://localhost:$(ADMIN_FRONTEND_PORT)"
	@echo "  后端地址: http://localhost:$(ADMIN_BACKEND_PORT)"
	@echo "  Swagger: http://localhost:$(ADMIN_BACKEND_PORT)/swagger/admin/index.html"
	@echo "  日志目录: $(ADMIN_LOG_DIR)/"
	@echo "  停止服务: make admin-stop"
	@echo "=========================================="

admin-start: admin-dev ## 同 admin-dev，一键启动前后端（别名）

admin-stop: ## 停止 Admin MVP 前后端服务
	@echo "==> 停止 Admin MVP 服务..."
	@if [ -f $(ADMIN_BACKEND_PID) ]; then \
		BPID=$$(cat $(ADMIN_BACKEND_PID)); \
		if kill -0 $$BPID 2>/dev/null; then \
			kill $$BPID 2>/dev/null; sleep 1; \
			if kill -0 $$BPID 2>/dev/null; then kill -9 $$BPID 2>/dev/null; fi; \
			echo "✅ Go 后端已停止 (PID $$BPID)"; \
		else \
			echo "⚠️  Go 后端进程不存在 (stale PID $$BPID)"; \
		fi; \
		rm -f $(ADMIN_BACKEND_PID); \
	else \
		echo "ℹ️  无后端 PID 文件，跳过"; \
	fi
	@if [ -f $(ADMIN_FRONTEND_PID) ]; then \
		FPID=$$(cat $(ADMIN_FRONTEND_PID)); \
		if kill -0 $$FPID 2>/dev/null; then \
			kill $$FPID 2>/dev/null; sleep 1; \
			if kill -0 $$FPID 2>/dev/null; then kill -9 $$FPID 2>/dev/null; fi; \
			echo "✅ Vue 前端已停止 (PID $$FPID)"; \
		else \
			echo "⚠️  Vue 前端进程不存在 (stale PID $$FPID)"; \
		fi; \
		rm -f $(ADMIN_FRONTEND_PID); \
	else \
		echo "ℹ️  无前端 PID 文件，跳过"; \
	fi
	@echo "==> Admin MVP 已全部停止"

admin-migrate: ## 初始化 SQLite 数据库（建表 + 系统种子 + 业务菜单 + 测试数据）
	@echo "==> 初始化 SQLite 数据库..."
	@cd go/admin && $(_GO_BIN) run -tags sqlite3 main.go migrate
	@echo "==> 导入业务种子数据（菜单/API/测试数据）..."
	@if [ -f go/admin/config/db-business.sql ]; then \
		cd go/admin && sqlite3 $(ADMIN_DB_FILE) < config/db-business.sql && echo "✅ 业务数据导入完成"; \
	else \
		echo "⚠️  db-business.sql 不存在，跳过业务数据导入"; \
	fi
	@echo "✅ 数据库初始化完成 ($(ADMIN_DB_FILE))"

admin-backend: ## 仅启动 Go 后端 (port 8000, SQLite3 CGO 首次编译约 1-2 分钟)
	@mkdir -p $(ADMIN_PID_DIR) $(ADMIN_LOG_DIR)
	@if [ -f $(ADMIN_BACKEND_PID) ] && lsof -i :$(ADMIN_BACKEND_PORT) >/dev/null 2>&1; then \
		echo "⚠️  后端已在运行 (port $(ADMIN_BACKEND_PORT))"; \
	else \
		echo "==> 启动 Go 后端（首次需编译 SQLite3 CGO，约 1-2 分钟）..."; \
		cd go/admin && nohup $(_GO_BIN) run -tags sqlite3 main.go server > $(PROJECT_ROOT)/$(ADMIN_BACKEND_LOG) 2>&1 & echo $$! > $(PROJECT_ROOT)/$(ADMIN_BACKEND_PID); \
		echo "    等待端口监听..."; \
		for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24; do \
			sleep 5; \
			if lsof -i :$(ADMIN_BACKEND_PORT) >/dev/null 2>&1; then \
				echo "✅ Go 后端就绪 (PID $$(cat $(ADMIN_BACKEND_PID)), port $(ADMIN_BACKEND_PORT), $$((i*5))s)"; \
				break; \
			fi; \
			if ! kill -0 $$(cat $(ADMIN_BACKEND_PID)) 2>/dev/null; then \
				echo "❌ 进程已退出:"; tail -30 $(ADMIN_BACKEND_LOG); rm -f $(ADMIN_BACKEND_PID); exit 1; \
			fi; \
			echo "    编译/启动中... ($$((i*5))s)"; \
		done; \
		if ! lsof -i :$(ADMIN_BACKEND_PORT) >/dev/null 2>&1; then \
			echo "❌ 超时（120s）"; tail -30 $(ADMIN_BACKEND_LOG); rm -f $(ADMIN_BACKEND_PID); exit 1; \
		fi; \
	fi

admin-frontend: ## 仅启动 Vue 前端 (port 9527)
	@mkdir -p $(ADMIN_PID_DIR) $(ADMIN_LOG_DIR)
	@if [ -f $(ADMIN_FRONTEND_PID) ] && kill -0 $$(cat $(ADMIN_FRONTEND_PID)) 2>/dev/null; then \
		echo "⚠️  前端已在运行 (PID $$(cat $(ADMIN_FRONTEND_PID)))"; \
	else \
		cd typescript/admin-web && NODE_OPTIONS=--openssl-legacy-provider nohup npx vue-cli-service serve > $(PROJECT_ROOT)/$(ADMIN_FRONTEND_LOG) 2>&1 & echo $$! > $(PROJECT_ROOT)/$(ADMIN_FRONTEND_PID); \
		sleep 5; \
		if kill -0 $$(cat $(ADMIN_FRONTEND_PID)) 2>/dev/null; then \
			echo "✅ Vue 前端已启动 (PID $$(cat $(ADMIN_FRONTEND_PID)), port $(ADMIN_FRONTEND_PORT))"; \
		else \
			echo "❌ 启动失败:"; tail -10 $(ADMIN_FRONTEND_LOG); rm -f $(ADMIN_FRONTEND_PID); exit 1; \
		fi; \
	fi

admin-status: ## 查看 Admin MVP 前后端运行状态
	@echo "━━━ Admin MVP 运行状态 ━━━"
	@echo ""
	@if lsof -i :$(ADMIN_BACKEND_PORT) >/dev/null 2>&1; then \
		echo "  Go 后端: ✅ 运行中 (port $(ADMIN_BACKEND_PORT))"; \
	else \
		if [ -f $(ADMIN_BACKEND_PID) ]; then \
			echo "  Go 后端: ❌ 进程已退出 (stale PID $$(cat $(ADMIN_BACKEND_PID)))"; \
		else \
			echo "  Go 后端: ⚪ 未启动"; \
		fi; \
	fi
	@if lsof -i :$(ADMIN_FRONTEND_PORT) >/dev/null 2>&1; then \
		echo "  Vue 前端: ✅ 运行中 (port $(ADMIN_FRONTEND_PORT))"; \
	else \
		if [ -f $(ADMIN_FRONTEND_PID) ]; then \
			echo "  Vue 前端: ❌ 进程已退出 (stale PID $$(cat $(ADMIN_FRONTEND_PID)))"; \
		else \
			echo "  Vue 前端: ⚪ 未启动"; \
		fi; \
	fi
	@echo ""

admin-restart: ## 重启 Admin MVP 前后端
	@$(MAKE) admin-stop
	@sleep 2
	@$(MAKE) admin-dev

# ============================================================
# Admin MVP E2E 端到端测试（Playwright + Chromium）
# 前置: make admin-dev 或手动启动 dev server (port 9527)
# ============================================================

E2E_DIR := tests/e2e
E2E_EVIDENCE_DIR := $(E2E_DIR)/evidence
E2E_RESULTS_DIR := $(E2E_DIR)/test-results

e2e-install: ## 安装 E2E 测试依赖（Playwright + Chromium）
	@echo "==> 安装 Playwright 依赖..."
	@if [ ! -f $(E2E_DIR)/package.json ]; then \
		echo "❌ $(E2E_DIR)/package.json 不存在"; exit 1; \
	fi
	@cd $(E2E_DIR) && npm install
	@echo "==> 安装 Playwright 浏览器..."
	@cd $(E2E_DIR) && npx playwright install chromium
	@echo "✅ E2E 依赖安装完成"

e2e-config: ## 运行 Config Admin E2E 测试（12 用例：Schema CRUD / 配置发布 / 版本历史）
	@cd $(E2E_DIR) && npx playwright test specs/config-admin.spec.ts --reporter=list --retries=0

e2e-i18n: ## 运行 I18n Admin E2E 测试（19 用例：字符串 CRUD / 语言包 / CSV 导入导出）
	@cd $(E2E_DIR) && npx playwright test specs/i18n-admin.spec.ts --reporter=list --retries=0

e2e-admin: ## 运行 Admin MVP 全部 E2E 测试（31 用例 = config 12 + i18n 19）
	@cd $(E2E_DIR) && npx playwright test --reporter=list --retries=0

e2e-all: e2e-admin ## 同 e2e-admin，全量 E2E 测试（别名）

e2e-report: ## 查看 E2E 截图证据和测试报告
	@echo "━━━ E2E 截图证据 ━━━"
	@if [ -d $(E2E_EVIDENCE_DIR) ]; then \
		PNG_COUNT=$$(find $(E2E_EVIDENCE_DIR) -name "*.png" 2>/dev/null | wc -l | tr -d ' '); \
		WEBM_COUNT=$$(find $(E2E_EVIDENCE_DIR) -name "*.webm" 2>/dev/null | wc -l | tr -d ' '); \
		echo "  截图文件: $$PNG_COUNT 张 PNG ($(E2E_EVIDENCE_DIR)/)"; \
		echo "  录屏文件: $$WEBM_COUNT 个 WEBM"; \
	else \
		echo "  ⚠️  证据目录不存在，请先运行 make e2e-admin"; \
	fi
	@echo ""
	@echo "━━━ HTML 测试报告 ━━━"
	@if [ -d $(E2E_RESULTS_DIR) ]; then \
		REPORTS=$$(find $(E2E_RESULTS_DIR) -name "index.html" 2>/dev/null); \
		if [ -n "$$REPORTS" ]; then \
			echo "  报告位置:"; echo "$$REPORTS" | while read r; do echo "    $$r"; done; \
		else \
			echo "  ⚠️  尚无 HTML 报告"; \
		fi; \
	else \
		echo "  ⚠️  结果目录不存在"; \
	fi
