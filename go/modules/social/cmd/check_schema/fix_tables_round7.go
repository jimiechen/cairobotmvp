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

	// Round 7: 清理残留的空 UID 用户（之前 SvcRegister 未生成 UID 导致）
	sqls := []string{
		"DELETE FROM `member_stats` WHERE `user_id` IN (SELECT `id` FROM `users` WHERE `uid` = '' AND `id` LIKE 'user-%')",
		"DELETE FROM `users` WHERE `uid` = '' AND `id` LIKE 'user-%'",
	}

	for i, sql := range sqls {
		result := db.Exec(sql)
		if result.Error != nil {
			fmt.Printf("[%d] SKIP: %v\n", i+1, result.Error.Error())
		} else {
			fmt.Printf("[%d] OK: rows=%d\n", i+1, result.RowsAffected)
		}
	}
	fmt.Println("\nRound 7 完成: 清理空 UID 残留用户")
}
