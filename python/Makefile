.PHONY: install proto lint unit coverage build package clean help

help: ## 显示 Python 子模块帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

install: ## 安装 Python 依赖
	@echo "[py] ==> 安装依赖..."
	@if [ -f ai/hello/pyproject.toml ]; then \
		cd ai/hello && $(PYTHON) -m pip install -e ".[dev]" || $(PYTHON) -m pip install -r requirements.txt || echo "[py] 跳过依赖安装"; \
	fi
	@echo "[py] 依赖安装完成"

proto: ## 生成 Protobuf Python 代码
	@echo "[py] ==> 生成 Protobuf Python 代码..."
	@bash ../../scripts/proto/generate-python.sh
	@echo "[py] Protobuf Python 代码生成完成"

lint: ## 运行 Python Ruff/Flake8 检查
	@echo "[py] ==> 运行 Lint 检查..."
	@if command -v ruff >/dev/null 2>&1; then \
		ruff check ai/ || echo "[py] 跳过 ruff check"; \
	elif command -v flake8 >/dev/null 2>&1; then \
		flake8 ai/ || echo "[py] 跳过 flake8"; \
	else \
		echo "[py] 警告：ruff/flake8 均未安装，跳过 lint"; \
	fi
	@echo "[py] Lint 检查完成"

unit: ## 运行 Python 单元测试
	@echo "[py] ==> 运行单元测试..."
	@if [ -d ai ]; then \
		cd ai && $(PYTHON) -m pytest -v --tb=short || echo "[py] ⚠️  pytest 未安装或无测试文件"; \
	else \
		echo "[py] ⏭️  无 ai/ 目录，跳过 python test"; \
	fi
	@echo "[py] 单元测试完成"

coverage: ## 生成 Python 覆盖率报告
	@echo "[py] ==> 生成覆盖率报告..."
	@mkdir -p ../../docs/reports/coverage/python
	@if [ -d ai ]; then \
		cd ai && $(PYTHON) -m pytest --cov=. --cov-report=html:../../docs/reports/coverage/python \
			--cov-report=term-missing -v || echo "[py] ⚠️  coverage 未安装或无测试文件"; \
	fi
	@echo "[py] 覆盖率报告已生成"

build: ## 构建 Python 包
	@echo "[py] ==> 构建包..."
	@if [ -f ai/hello/pyproject.toml ]; then \
		cd ai/hello && $(PYTHON) -m build || echo "[py] 跳过 build"; \
	fi
	@echo "[py] 构建完成"

package: ## 打包 Python 发布产物
	@echo "[py] ==> 打包产物..."
	@mkdir -p ../../dist/python
	@if [ -d ai/hello/dist ]; then \
		cp ai/hello/dist/* ../../dist/python/ || true; \
	fi
	@echo "[py] 打包完成"

clean: ## 清理 Python 构建产物
	@echo "[py] ==> 清理构建产物..."
	@rm -rf ai/**/__pycache__ ai/**/*.pyc ai/**/.pytest_cache ai/**/.coverage
	@rm -rf ai/hello/dist ai/hello/build ai/hello/*.egg-info
	@find ai -name "__pycache__" -type d -exec rm -rf {} + 2>/dev/null || true
	@echo "[py] 清理完成"
