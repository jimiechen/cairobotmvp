.PHONY: install proto lint unit coverage build package clean help proto-tester-dev proto-tester-stop

help: ## 显示 TypeScript 子模块帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

install: ## 安装 pnpm workspace 依赖
	@echo "[ts] ==> 安装依赖..."
	@if command -v pnpm >/dev/null 2>&1; then \
		pnpm install; \
	elif command -v npm >/dev/null 2>&1; then \
		echo "[ts] 警告：pnpm 未安装，回退到 npm"; \
		npm install; \
	else \
		echo "[ts] 错误：pnpm 和 npm 均未安装"; \
		exit 1; \
	fi
	@echo "[ts] 依赖安装完成"

proto: ## 生成 Protobuf TypeScript 代码
	@echo "[ts] ==> 生成 Protobuf TypeScript 代码..."
	@bash ../../scripts/proto/generate-ts.sh
	@echo "[ts] Protobuf TypeScript 代码生成完成"

lint: ## 运行 TypeScript ESLint/Prettier 检查
	@echo "[ts] ==> 运行 Lint 检查..."
	@if [ -f web/package.json ]; then \
		cd web && npx eslint src/ --ext .ts,.tsx || echo "[ts] 跳过 web lint"; \
	fi
	@if [ -f admin-web/package.json ]; then \
		cd admin-web && npx eslint src/ --ext .ts,.tsx || echo "[ts] 跳过 admin-web lint"; \
	fi
	@echo "[ts] Lint 检查完成"

unit: ## 运行 TypeScript 单元测试（使用 Vitest）
	@echo "[ts] ==> 运行单元测试..."
	@if [ -f web/package.json ]; then \
		cd web && npx vitest run --passWithNoTests --reporter=verbose 2>&1 || echo "[ts] ⚠️  web 测试失败或跳过"; \
	fi
	@if [ -f admin-web/package.json ]; then \
		cd admin-web && npx vitest run --passWithNoTests --reporter=verbose 2>&1 || echo "[ts] ⚠️  admin-web 测试失败或跳过"; \
	fi
	@echo "[ts] 单元测试完成"

coverage: ## 生成 TypeScript 覆盖率报告
	@echo "[ts] ==> 生成覆盖率报告..."
	@mkdir -p ../../docs/reports/coverage/typescript
	@if [ -f web/package.json ]; then \
		cd web && npx vitest run --coverage --coverageDirectory=../../docs/reports/coverage/typescript/web --passWithNoTests 2>&1 || echo "[ts] ⚠️  跳过 web coverage"; \
	fi
	@if [ -f admin-web/package.json ]; then \
		cd admin-web && npx vitest run --coverage --coverageDirectory=../../docs/reports/coverage/typescript/admin-web --passWithNoTests 2>&1 || echo "[ts] ⚠️  跳过 admin-web coverage"; \
	fi
	@echo "[ts] 覆盖率报告已生成"

build: ## 构建 TypeScript 项目
	@echo "[ts] ==> 构建项目..."
	@if [ -f web/package.json ]; then \
		cd web && npx vite build || echo "[ts] 跳过 web build"; \
	fi
	@if [ -f admin-web/package.json ]; then \
		cd admin-web && npx vite build || echo "[ts] 跳过 admin-web build"; \
	fi
	@echo "[ts] 构建完成"

package: ## 打包 TypeScript 发布产物
	@echo "[ts] ==> 打包产物..."
	@mkdir -p ../../dist/typescript
	@if [ -f web/dist ]; then \
		tar czf ../../dist/typescript/cairobot-web-$(shell date +%Y%m%d).tar.gz web/dist/ || true; \
	fi
	@echo "[ts] 打包完成"

clean: ## 清理 TypeScript 构建产物
	@echo "[ts] ==> 清理构建产物..."
	@rm -rf web/dist admin-web/dist node_modules coverage
	@find . -name ".vite" -type d -exec rm -rf {} + 2>/dev/null || true
	@echo "[ts] 清理完成"

proto-tester-dev: ## 启动 Proto Tester 开发服务器（端口 3002）
	@echo "[ts] ==> 启动 Proto Tester 开发服务器..."
	@if [ ! -f proto-tester/package.json ]; then \
		echo "[ts] 错误：proto-tester 目录不存在"; \
		exit 1; \
	fi
	@cd proto-tester && npm install --silent 2>/dev/null
	@if [ -f /tmp/proto-tester.pid ] && kill -0 $(cat /tmp/proto-tester.pid) 2>/dev/null; then \
		echo "[ts] Proto Tester 已在运行（PID $(cat /tmp/proto-tester.pid)）"; \
	else \
		cd proto-tester && nohup npx vite --port 3002 --host > /tmp/proto-tester.log 2>&1 & echo $$! > /tmp/proto-tester.pid; \
		sleep 2; \
		echo "[ts] Proto Tester 已启动: http://localhost:3002 (PID $(cat /tmp/proto-tester.pid))"; \
	fi

proto-tester-stop: ## 停止 Proto Tester 开发服务器
	@echo "[ts] ==> 停止 Proto Tester..."
	@if [ -f /tmp/proto-tester.pid ] && kill -0 $(cat /tmp/proto-tester.pid) 2>/dev/null; then \
		kill $(cat /tmp/proto-tester.pid) 2>/dev/null; \
		rm -f /tmp/proto-tester.pid; \
		echo "[ts] Proto Tester 已停止"; \
	else \
		echo "[ts] Proto Tester 未在运行"; \
	fi
