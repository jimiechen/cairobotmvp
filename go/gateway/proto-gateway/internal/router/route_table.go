package router

import (
	"fmt"

	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/config"
)

// RouteTable 负责根据 maxType:minType 查找路由
// 命名为 RouteTable 而非 Router，避免与包名 router 产生歧义
type RouteTable struct {
	routes map[string]config.Route
}

// NewRouteTable 从配置创建 RouteTable
func NewRouteTable(cfg *config.RoutesConfig) *RouteTable {
	routes := make(map[string]config.Route, len(cfg.Routes))
	for _, r := range cfg.Routes {
		key := fmt.Sprintf("%d:%d", r.RequestMax, r.RequestMin)
		routes[key] = r
	}
	return &RouteTable{routes: routes}
}

// FindRoute 根据 maxType 和 minType 查找路由
func (rt *RouteTable) FindRoute(maxType int32, minType int32) (config.Route, bool) {
	key := fmt.Sprintf("%d:%d", maxType, minType)
	r, ok := rt.routes[key]
	return r, ok
}
