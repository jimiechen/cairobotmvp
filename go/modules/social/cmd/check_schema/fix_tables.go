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
		"ALTER TABLE `users` ADD COLUMN `uid` VARCHAR(20) DEFAULT NULL AFTER `id`",
		"ALTER TABLE `users` ADD UNIQUE INDEX `idx_uid` (`uid`)",
		"ALTER TABLE `member_stats` ADD COLUMN `id` CHAR(32) NOT NULL DEFAULT '' FIRST",
		"ALTER TABLE `member_stats` ADD PRIMARY KEY (`id`)",
	}

	for i, sql := range sqls {
		result := db.Exec(sql)
		if result.Error != nil {
			fmt.Printf("[%d] SKIP (可能已存在): %v\n", i+1, result.Error.Error())
		} else {
			fmt.Printf("[%d] OK: rows=%d\n", i+1, result.RowsAffected)
		}
	}
	fmt.Println("\n表结构修复完成")
}
