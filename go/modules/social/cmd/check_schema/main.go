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
		fmt.Println("ERR:", err)
		return
	}

	printCols := func(table string) {
		var cols []string
		db.Raw("SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=? AND TABLE_NAME=? ORDER BY ORDINAL_POSITION",
			os.Getenv("MYSQL_DATABASE"), table).Scan(&cols)
		fmt.Printf("=== %s (%d columns) ===\n", table, len(cols))
		for _, c := range cols {
			fmt.Printf("  %s\n", c)
		}
	}

	printCols("users")
	fmt.Println()
	printCols("member_stats")
	fmt.Println()
	printCols("member_blocks")
	fmt.Println()

	var tables []string
	db.Raw("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA=? AND (TABLE_NAME LIKE 'member%' OR TABLE_NAME LIKE 'user%' OR TABLE_NAME LIKE 'group%' OR TABLE_NAME LIKE 'topic%') ORDER BY TABLE_NAME",
		os.Getenv("MYSQL_DATABASE")).Scan(&tables)
	fmt.Println("=== all social tables ===")
	for _, t := range tables {
		fmt.Printf("  %s\n", t)
	}
}
