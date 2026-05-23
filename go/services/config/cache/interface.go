package cache

// Cache 配置缓存抽象接口
// 解耦 Service 层与具体缓存实现（Redis / 内存 / 伪实现）
// 键格式约定: cfg:{env}:{module_key} 用于版本缓存
type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Delete(key string)
	Invalidate(prefix string)
}
