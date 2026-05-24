# CaiRobot Go 环境专属配置
#
# 本目录下的配置文件仅适用于 Go 服务运行时环境（Tars Servant / Admin Server / Gateway）
# TypeScript/Python 前端不使用此配置。
#
# 使用方式:
#   Gateway 启动时读取: configs/go/gateway.yaml
#   Config/I18n Servant 启动时通过环境变量覆盖 (CONFIG_DB_PATH, I18N_DB_PATH)
#   Admin Server 启动时通过环境变量覆盖 (ADMIN_PORT, ADMIN_DB_PATH)

## gateway.yaml — Gateway + Tars 路由配置

gateway/routes.yaml:
  - 路由定义表（2100 段系统路由 + 6000 段配置/i18n 路由）
  - 由 Gateway cmd/server/main.go 启动时加载

## server.yaml — 共享基础设施配置（MySQL / Redis / Cache TTL）

server.yaml:
  - MySQL 连接配置（host/port/user/password/database）
  - Redis 连接配置（host/port/password/db）
  - Cache TTL 配置（config_ttl_seconds=30, i18n_ttl_seconds=60）
  - PubSub 开关
  - 注意: 当前 MVP 阶段使用 SQLite 作为默认存储，此 MySQL/Redis 配置为阶段 D 预留
