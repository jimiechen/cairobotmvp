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

	// Round 5: member_blocks.blocked_user_id 允许 NULL
	// Go Model (MemberBlock) 使用 blocked_id 字段，不映射 DB 的 blocked_user_id
	// GORM INSERT 不包含 blocked_user_id → NOT NULL 约束违反
	sqls := []string{
		"ALTER TABLE `member_blocks` MODIFY COLUMN `blocked_user_id` CHAR(32) DEFAULT NULL",
	}

	for i, sql := range sqls {
		result := db.Exec(sql)
		if result.Error != nil {
			fmt.Printf("[%d] SKIP: %v\n", i+1, result.Error.Error())
		} else {
			fmt.Printf("[%d] OK: rows=%d\n", i+1, result.RowsAffected)
		}
	}
	fmt.Println("\nRound 5 修复完成: blocked_user_id -> DEFAULT NULL")
}
