package main

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("连接失败:", err)
		os.Exit(1)
	}

	sqls := []string{
		// member_blocks: 添加 Model 需要的列
		"ALTER TABLE `member_blocks` ADD COLUMN `blocker_id` CHAR(32) DEFAULT NULL AFTER `id`",
		"ALTER TABLE `member_blocks` ADD COLUMN `blocked_id` CHAR(32) DEFAULT NULL AFTER `blocker_id`",
		"ALTER TABLE `member_blocks` ADD INDEX `idx_blocker_id` (`blocker_id`)",
		"ALTER TABLE `member_blocks` ADD INDEX `idx_blocked_id` (`blocked_id`)",
		// member_stats: 添加 Model 需要的统计列
		"ALTER TABLE `member_stats` ADD COLUMN `replies_count` INT NOT NULL DEFAULT 0 AFTER `topics_count`",
		"ALTER TABLE `member_stats` ADD COLUMN `likes_received` INT NOT NULL DEFAULT 0 AFTER `replies_count`",
		"ALTER TABLE `member_stats` ADD COLUMN `groups_joined` INT NOT NULL DEFAULT 0 AFTER `likes_received`",
	}

	for i, sql := range sqls {
		result := db.Exec(sql)
		if result.Error != nil {
			fmt.Printf("[%d] SKIP: %v\n", i+1, result.Error.Error())
		} else {
			fmt.Printf("[%d] OK: rows=%d\n", i+1, result.RowsAffected)
		}
	}
	fmt.Println("\nRound 2 表结构修复完成")
}
