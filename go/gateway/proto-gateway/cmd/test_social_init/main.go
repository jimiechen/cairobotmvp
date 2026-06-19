package main

import (
	"fmt"
	"os"

	"github.com/jimiechen/mineplanet/go/modules/social/member"
	"github.com/jimiechen/mineplanet/go/modules/social/group"
	"github.com/jimiechen/mineplanet/go/modules/social/topic"
	socialmodule "github.com/jimiechen/mineplanet/go/modules/social"
)

// 最小测试：验证 Social 模块初始化是否会导致 panic
func main() {
	fmt.Println("=== Social Module Init Test ===")

	// 1. 测试 MemoryRepository 创建
	fmt.Println("1. Creating MemoryRepositories...")
	memberRepo := member.NewMemoryRepository()
	groupRepo := group.NewMemoryRepository()
	topicRepo := topic.NewMemoryRepository()
	fmt.Println("   ✅ MemoryRepositories created")

	// 2. 测试 Social Module 创建
	fmt.Println("2. Creating SocialModule...")
	socialMod := socialmodule.NewModule(memberRepo, groupRepo, topicRepo)
	if socialMod == nil {
		fmt.Println("   ❌ SocialModule is nil")
		os.Exit(1)
	}
	fmt.Println("   ✅ SocialModule created")

	// 3. 测试 Servant 获取
	fmt.Println("3. Getting Servants...")
	if socialMod.MemberServant == nil {
		fmt.Println("   ❌ MemberServant is nil")
		os.Exit(1)
	}
	if socialMod.GroupServant == nil {
		fmt.Println("   ❌ GroupServant is nil")
		os.Exit(1)
	}
	if socialMod.TopicServant == nil {
		fmt.Println("   ❌ TopicServant is nil")
		os.Exit(1)
	}
	fmt.Println("   ✅ All Servants available")

	fmt.Println("\n=== All Tests Passed ===")
}
