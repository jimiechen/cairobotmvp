// Package cache 社交域缓存 Key 管理与失效接口
// 定义所有缓存 key 的构造函数和 CacheInvalidator 接口
// 缓存 key 必须通过本包函数生成，禁止手写拼接散落
//
// 设计依据：
//   - PRD-social-app-mvp.md §11.4 缓存 Key 命名规范总表
//   - ADR-social-data-level-and-cache-strategy.md 第四章缓存策略
package cache

import "fmt"

// ===== 成员域缓存 Key =====

// MemberProfileKey 用户资料缓存 key
func MemberProfileKey(userID string) string {
	return fmt.Sprintf("member:profile:%s", userID)
}

// MemberStatsKey 用户统计缓存 key
func MemberStatsKey(userID string) string {
	return fmt.Sprintf("member:stats:%s", userID)
}

// ===== 群组域缓存 Key =====

// GroupDetailKey 群组详情缓存 key
func GroupDetailKey(groupID string) string {
	return fmt.Sprintf("group:detail:%s", groupID)
}

// GroupMemberKey 成员关系缓存 key
func GroupMemberKey(groupID, userID string) string {
	return fmt.Sprintf("group:member:%s:%s", groupID, userID)
}

// GroupMembersKey 成员列表缓存 key（分页）
func GroupMembersKey(groupID string, page int) string {
	return fmt.Sprintf("group:members:%s:%d", groupID, page)
}

// GroupPlansKey 付费方案列表缓存 key
func GroupPlansKey(groupID string) string {
	return fmt.Sprintf("group:plans:%s", groupID)
}

// GroupStatsKey 群组统计缓存 key
func GroupStatsKey(groupID string) string {
	return fmt.Sprintf("group:stats:%s", groupID)
}

// OwnerDashboardKey 圈主看板缓存 key
func OwnerDashboardKey(ownerID string) string {
	return fmt.Sprintf("owner:dashboard:%s", ownerID)
}

// GroupRecommendKey 推荐群组缓存 key
func GroupRecommendKey(category string) string {
	return fmt.Sprintf("group:recommend:%s", category)
}

// ===== 主题域缓存 Key =====

// TopicDetailKey 帖子详情缓存 key
func TopicDetailKey(topicID string) string {
	return fmt.Sprintf("topic:detail:%s", topicID)
}

// TopicStatsKey 帖子统计缓存 key
func TopicStatsKey(topicID string) string {
	return fmt.Sprintf("topic:stats:%s", topicID)
}

// TopicReadKey 阅读记录缓存 key
func TopicReadKey(userID, topicID string) string {
	return fmt.Sprintf("topic:read:%s:%s", userID, topicID)
}

// GroupTopicsKey 群组帖子列表缓存 key（分页）
func GroupTopicsKey(groupID string, page int) string {
	return fmt.Sprintf("group:topics:%s:%d", groupID, page)
}

// TopicHotKey 热门帖子排行缓存 key
func TopicHotKey(groupID string) string {
	return fmt.Sprintf("topic:hot:%s", groupID)
}
