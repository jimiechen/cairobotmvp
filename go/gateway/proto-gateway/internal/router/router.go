package router

import (
	"fmt"

	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/config"
)

// Router 负责根据 maxType:minType 查找路由
type Router struct {
	routes map[string]config.Route
}

// NewRouter 从配置创建 Router
func NewRouter(cfg *config.RoutesConfig) *Router {
	routes := make(map[string]config.Route, len(cfg.Routes))
	for _, r := range cfg.Routes {
		key := fmt.Sprintf("%d:%d", r.RequestMax, r.RequestMin)
		routes[key] = r
	}
	return &Router{routes: routes}
}

// FindRoute 根据 maxType 和 minType 查找路由
func (rt *Router) FindRoute(maxType int32, minType int32) (config.Route, bool) {
	key := fmt.Sprintf("%d:%d", maxType, minType)
	r, ok := rt.routes[key]
	return r, ok
}
