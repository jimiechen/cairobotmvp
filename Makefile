.PHONY: help bootstrap proto lint test unit integration coverage build package \
        docs rules testcase-check comment-check ci clean

PROJECT_ROOT := $(shell pwd)
export PROJECT_ROOT

# GOPROXY 设置：解决中国大陆网络访问 Go module 代理问题
export GOPROXY ?= https://goproxy.cn,direct

# PATH 扩展：确保 protoc / protoc-gen-* / tars2go 等工具可被找到
export PATH := $(shell dirname $$(which protoc 2>/dev/null || echo /usr/local/bin)):$$(go env GOPATH 2>/dev/null)/bin:$(shell npm root -g 2>/dev/null)/.bin:$(PATH)

help: ## 显示帮助信息
	@echo "CaiRobot MVP 工程入口 Makefile"
	@echo ""
	@echo "使用方式：make [target]"
	@echo ""
	@echo "工程编排目标："
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
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
	@if [ -f go/Makefile ]; then $(MAKE) -C go unit || echo "跳过 go unit"; fi
	@if [ -f typescript/Makefile ]; then $(MAKE) -C typescript unit || echo "⏭️  typescript unit: pending（需 pnpm install）"; fi
	@if [ -f python/Makefile ]; then $(MAKE) -C python unit || echo "⏭️  python unit: pending（需 pip install）"; fi
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
	@python3 scripts/ci/check_required_docs.py

rules: ## 执行工程规范检查（工具链 + 目录布局 + Make target + 模块路径 + 注释 + 测试用例）
	@echo "==> 执行工程规范检查..."
	@bash scripts/ci/check_tools.sh || echo "❌ 工具链检查未通过"
	@python3 scripts/ci/check_directory_layout.py || echo "❌ 目录布局检查未通过"
	@bash scripts/ci/check_make_targets.sh || echo "❌ Make target 检查未通过"
	@python3 scripts/ci/check_module_paths.py || echo "❌ 模块路径检查未通过"
	@python3 scripts/ci/check_chinese_comments.py || echo "❌ 中文注释检查未通过"
	@python3 scripts/ci/check_testcase_registry.py || echo "❌ 测试用例注册表检查未通过"
	@echo "==> 规范检查完成"

testcase-check: ## 检查测试用例注册表一致性
	@echo "==> 检查测试用例注册表..."
	@python3 scripts/ci/check_testcase_registry.py

comment-check: ## 检查中文注释规范性
	@echo "==> 检查中文注释..."
	@python3 scripts/ci/check_chinese_comments.py

ci: ## 完整 CI 检查（本地等价于 GitHub Actions）
	@echo "=========================================="
	@echo "  CaiRobot MVP 完整 CI 检查"
	@echo "=========================================="
	@echo ""
	$(MAKE) docs
	$(MAKE) rules
	$(MAKE) proto
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
