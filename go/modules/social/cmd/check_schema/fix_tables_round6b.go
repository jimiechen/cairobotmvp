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

	// Round 6b: 清理残留测试数据（stats 按 user_id 匹配）
	sqls := []string{
		"DELETE FROM `member_stats` WHERE `user_id` LIKE 'gt_mbr_%'",
		"DELETE FROM `member_blocks` WHERE `id` LIKE 'gt_mbr_%'",
		"DELETE FROM `users` WHERE `id` LIKE 'gt_mbr_%'",
	}

	for i, sql := range sqls {
		result := db.Exec(sql)
		if result.Error != nil {
			fmt.Printf("[%d] SKIP: %v\n", i+1, result.Error.Error())
		} else {
			fmt.Printf("[%d] OK: rows=%d\n", i+1, result.RowsAffected)
		}
	}
	fmt.Println("\nRound 6b 完成: 残留数据清理（user_id 匹配）")
}
