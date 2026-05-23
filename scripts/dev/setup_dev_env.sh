#!/bin/bash
# ============================================================
# CaiRobot MVP 开发环境一键配置脚本
# 功能：自动检测并配置 Go/Node.js/Python/Protobuf 环境变量
# 使用方式：source scripts/dev/setup_dev_env.sh
# 注意：必须使用 source 或 . 执行，否则环境变量不会生效
# ============================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}  CaiRobot MVP 开发环境配置${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

# ============================================================
# 1. Go 环境配置
# ============================================================
setup_go_env() {
    echo -e "${YELLOW}[1/4] 配置 Go 环境...${NC}"
    
    local GO_BIN=""
    
    # 查找顺序：PATH → GOROOT → Homebrew → GOPATH → /usr/local/go
    if command -v go &> /dev/null; then
        GO_BIN=$(command -v go)
    elif [ -n "$GOROOT" ] && [ -x "$GOROOT/bin/go" ]; then
        GO_BIN="$GOROOT/bin/go"
    elif [ -x "/usr/local/opt/go/bin/go" ]; then
        GO_BIN="/usr/local/opt/go/bin/go"
    elif [ -x "$HOME/go/bin/go" ]; then
        GO_BIN="$HOME/go/bin/go"
    elif [ -x "/usr/local/go/bin/go" ]; then
        GO_BIN="/usr/local/go/bin/go"
    fi
    
    if [ -n "$GO_BIN" ]; then
        local GOROOT_PATH=$(dirname "$(dirname "$GO_BIN")")
        local DEFAULT_GOPATH=$("$GO_BIN" env GOPATH 2>/dev/null || echo "$HOME/go")
        local GOPATH_PATH="$DEFAULT_GOPATH"
        
        # 修复 GOPATH 与 GOROOT 冲突问题
        # 如果两者相同，自动将 GOPATH 迁移到独立的工作区目录
        if [ "$GOROOT_PATH" = "$GOPATH_PATH" ]; then
            echo -e "  ${YELLOW}⚠️  检测到 GOPATH 与 GOROOT 冲突${NC}"
            echo -e "     GOROOT 和 GOPATH 都指向: $GOROOT_PATH"
            
            # 尝试使用备用路径
            local ALTERNATIVE_GOPATH="$HOME/gowork"
            if [ -d "$ALTERNATIVE_GOPATH" ] || [ ! -d "$DEFAULT_GOPATH" ] || [ "$(ls -A "$DEFAULT_GOPATH/src" 2>/dev/null | wc -l)" -eq 0 ]; then
                GOPATH_PATH="$ALTERNATIVE_GOPATH"
                echo -e "     自动切换 GOPATH 到: $GOPATH_PATH"
            else
                # 如果原 GOPATH 有重要数据，使用更明确的路径
                GOPATH_PATH="$HOME/develop/gopath"
                echo -e "     自动切换 GOPATH 到: $GOPATH_PATH"
                echo -e "     ${YELLOW}提示: 原 GOPATH ($DEFAULT_GOPATH) 可能包含项目代码${NC}"
            fi
            
            # 创建新的 GOPATH 目录结构
            mkdir -p "$GOPATH_PATH"/{src,pkg,bin}
        fi
        
        export GOROOT="$GOROOT_PATH"
        export GOPATH="$GOPATH_PATH"
        export PATH="$GOPATH_PATH/bin:$GOROOT_PATH/bin:$PATH"
        export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
        
        echo -e "  ${GREEN}✅ Go 已找到${NC}"
        echo -e "     路径: $GO_BIN"
        echo -e "     版本: $(go version)"
        echo -e "     GOROOT: $GOROOT (Go 安装目录)"
        echo -e "     GOPATH: $GOPATH (工作区目录)"
        echo -e "     GOPROXY: $GOPROXY"
    else
        echo -e "  ${RED}❌ Go 未找到${NC}"
        echo -e "     请安装 Go: https://go.dev/doc/install"
        return 1
    fi
    
    echo ""
}

# ============================================================
# 2. Python 环境配置
# ============================================================
setup_python_env() {
    echo -e "${YELLOW}[2/4] 配置 Python 环境...${NC}"
    
    local PYTHON_BIN=""
    
    # 查找顺序：python3 → python → Homebrew → pyenv → conda → miniconda3 → anaconda3 → /usr/local/bin
    if command -v python3 &> /dev/null; then
        PYTHON_BIN=$(command -v python3)
    elif command -v python &> /dev/null; then
        PYTHON_BIN=$(command -v python)
    elif [ -x "/usr/local/opt/python/libexec/bin/python3" ]; then
        PYTHON_BIN="/usr/local/opt/python/libexec/bin/python3"
    elif [ -d "$HOME/.pyenv" ] && [ -x "$HOME/.pyenv/shims/python3" ]; then
        PYTHON_BIN=$(pyenv which python3 2>/dev/null || echo "")
    elif [ -x "$HOME/miniconda3/bin/python3" ]; then
        PYTHON_BIN="$HOME/miniconda3/bin/python3"
    elif [ -x "$HOME/anaconda3/bin/python3" ]; then
        PYTHON_BIN="$HOME/anaconda3/bin/python3"
    elif [ -x "/usr/local/bin/python3" ]; then
        PYTHON_BIN="/usr/local/bin/python3"
    fi
    
    if [ -n "$PYTHON_BIN" ]; then
        export PYTHON="$PYTHON_BIN"
        local PYTHON_DIR=$(dirname "$PYTHON_BIN")
        
        # 将 Python bin 目录加入 PATH
        export PATH="$PYTHON_DIR:${PYTHON_DIR}/../bin:$PATH"
        
        # PIP 使用国内镜像加速
        export PIP_INDEX_URL="${PIP_INDEX_URL:-https://pypi.tuna.tsinghua.edu.cn/simple}"
        export PIP_TRUSTED_HOST="${PIP_TRUSTED_HOST:-pypi.tuna.tsinghua.edu.cn}"
        
        echo -e "  ${GREEN}✅ Python 已找到${NC}"
        echo -e "     路径: $PYTHON_BIN"
        echo -e "     版本: $("$PYTHON_BIN" --version 2>&1)"
        echo -e "     PIP_INDEX_URL: $PIP_INDEX_URL"
    else
        echo -e "  ${RED}❌ Python 未找到${NC}"
        echo -e "     请安装 Python: https://www.python.org/downloads/"
        export PYTHON="python3"
        return 1
    fi
    
    echo ""
}

# ============================================================
# 3. Node.js 环境配置
# ============================================================
setup_node_env() {
    echo -e "${YELLOW}[3/4] 配置 Node.js 环境...${NC}"
    
    local NODE_BIN=""
    
    # 查找顺序：PATH → Homebrew → nvm → fnm → volta → /usr/local/bin
    if command -v node &> /dev/null; then
        NODE_BIN=$(command -v node)
    elif [ -x "/usr/local/opt/node/bin/node" ]; then
        NODE_BIN="/usr/local/opt/node/bin/node"
    elif [ -d "$HOME/.nvm" ]; then
        # 找到最新版本的 nvm node
        NODE_BIN=$(ls -t "$HOME/.nvm/versions/node/"*/bin/node 2>/dev/null | head -1)
    elif [ -d "$HOME/.fnm" ] && [ -x "$HOME/.fnm/current/bin/node" ]; then
        NODE_BIN="$HOME/.fnm/current/bin/node"
    elif [ -d "$HOME/.volta" ] && [ -x "$HOME/.volta/bin/node" ]; then
        NODE_BIN="$HOME/.volta/bin/node"
    elif [ -x "/usr/local/bin/node" ]; then
        NODE_BIN="/usr/local/bin/node"
    fi
    
    if [ -n "$NODE_BIN" ]; then
        export NODE="$NODE_BIN"
        local NODE_DIR=$(dirname "$NODE_BIN")
        
        # 将 Node bin 目录加入 PATH
        export PATH="$NODE_DIR:$PATH"
        
        # npm 使用国内镜像加速
        export NPM_REGISTRY="${NPM_REGISTRY:-https://registry.npmmirror.com}"
        export ELECTRON_MIRROR="${ELECTRON_MIRROR:-https://npmmirror.com/mirrors/electron/}"
        
        echo -e "  ${GREEN}✅ Node.js 已找到${NC}"
        echo -e "     路径: $NODE_BIN"
        echo -e "     版本: $("$NODE_BIN" --version 2>&1)"
        
        # 检查 npm
        if command -v npm &> /dev/null; then
            echo -e "     npm: $(npm --version 2>&1)"
        fi
        
        echo -e "     NPM_REGISTRY: $NPM_REGISTRY"
    else
        echo -e "  ${RED}❌ Node.js 未找到${NC}"
        echo -e "     请安装 Node.js: https://nodejs.org/"
        export NODE="node"
        return 1
    fi
    
    echo ""
}

# ============================================================
# 4. Protobuf 工具配置
# ============================================================
setup_protobuf_env() {
    echo -e "${YELLOW}[4/4] 配置 Protobuf 工具...${NC}"
    
    local PROTOC_BIN=""
    
    if command -v protoc &> /dev/null; then
        PROTOC_BIN=$(command -v protoc)
    fi
    
    if [ -n "$PROTOC_BIN" ]; then
        local PROTOC_DIR=$(dirname "$PROTOC_BIN")
        export PATH="$PROTOC_DIR:$PATH"
        
        echo -e "  ${GREEN}✅ protoc 已找到${NC}"
        echo -e "     路径: $PROTOC_BIN"
        echo -e "     版本: $(protoc --version 2>&1)"
    else
        echo -e "  ${YELLOW}⚠️  protoc 未找到（可选）${NC}"
        echo -e "     如需生成 Protobuf 代码，请安装:"
        echo -e "       macOS: brew install protobuf"
        echo -e "       Linux: apt install -y protobuf-compiler"
    fi
    
    echo ""
}

# ============================================================
# 主流程
# ============================================================
main() {
    # 设置项目根目录
    export PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
    
    echo -e "项目根目录: ${GREEN}$PROJECT_ROOT${NC}"
    echo ""
    
    # 配置各语言环境
    setup_go_env
    setup_python_env
    setup_node_env
    setup_protobuf_env
    
    # 输出总结
    echo -e "${BLUE}============================================${NC}"
    echo -e "${GREEN}  ✅ 开发环境配置完成${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo ""
    echo -e "使用提示:"
    echo -e "  1. 运行测试: make test"
    echo -e "  2. 初始化依赖: make bootstrap"
    echo -e "  3. 完整 CI 检查: make ci"
    echo -e "  4. 查看帮助: make help"
    echo ""
    echo -e "${YELLOW}注意: 此配置仅在当前终端会话中生效${NC}"
    echo -e "如需永久生效，请将以下内容添加到 ~/.zshrc:"
    echo -e "  source $PROJECT_ROOT/scripts/dev/setup_dev_env.sh"
    echo ""
}

# 执行主函数
main
