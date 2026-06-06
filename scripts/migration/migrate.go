// S1 MySQL 迁移执行器
//
// 用法：
//   cd scripts/migration && go run migrate.go 002_s1_mysql_tables.sql 003_s1_seed_data.sql
//
// 环境变量（可选，有默认值）：
//   MYSQL_HOST     默认 127.0.0.1
//   MYSQL_PORT     默认 3306
//   MYSQL_USER     默认 root
//   MYSQL_PASSWORD 默认空
//   MYSQL_DATABASE 默认 cairobot_db
package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	flag.Parse()
	files := flag.Args()

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "用法: go run migrate.go <sql文件> [sql文件...]")
		fmt.Fprintln(os.Stderr, "环境变量: MYSQL_HOST MYSQL_PORT MYSQL_USER MYSQL_PASSWORD")
		os.Exit(1)
	}

	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "root")
	pass := getEnv("MYSQL_PASSWORD", "")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "连接 MySQL 失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Ping MySQL 失败: %v\n", err)
		fmt.Fprintln(os.Stderr, "请确认 MySQL 服务已启动且连接参数正确")
		os.Exit(1)
	}

	fmt.Printf("已连接 MySQL %s:%s\n\n", host, port)

	for _, f := range files {
		if err := executeSQLFile(db, f); err != nil {
			fmt.Fprintf(os.Stderr, "\n执行 %s 失败: %v\n", f, err)
			os.Exit(1)
		}
	}

	fmt.Println("\n=== 所有迁移脚本执行完成 ===")
}

// executeSQLFile 读取并逐条执行 SQL 文件
func executeSQLFile(db *sql.DB, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取文件: %w", err)
	}

	fmt.Printf(">>> 执行: %s (%d bytes)\n", path, len(content))

	statements := splitStatements(string(content))
	successCount := 0

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			errMsg := err.Error()
			// CREATE DATABASE IF NOT EXISTS / INSERT IGNORE 等幂等操作在重复执行时的警告可忽略
			if strings.Contains(strings.ToLower(errMsg), "already exists") ||
				strings.Contains(strings.ToLower(errMsg), "duplicate") {
				fmt.Printf("    [%02d] SKIP (已存在): %s\n", i+1, truncate(stmt, 60))
				continue
			}
			return fmt.Errorf("语句[%02d] 执行失败: %w\n  SQL: %s", i+1, err, truncate(stmt, 120))
		}
		successCount++
		upper := strings.ToUpper(stmt)
		if strings.HasPrefix(upper, "CREATE") ||
			strings.HasPrefix(upper, "USE") ||
			strings.HasPrefix(upper, "INSERT") {
			fmt.Printf("    [%02d] OK: %s\n", i+1, truncate(stmt, 70))
		}
	}

	fmt.Printf("<<< 完成: %d 条语句成功\n", successCount)
	return nil
}

// ---- SQL 解析工具 ----

// splitStatements 按分号分割 SQL，跳过注释行
func splitStatements(sql string) []string {
	var stmts []string
	var current strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(sql))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 跳过单行注释
		if strings.HasPrefix(trimmed, "--") {
			if current.Len() > 0 {
				stmts = append(stmts, strings.TrimSpace(current.String()))
				current.Reset()
			}
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		// 以分号结尾的完整语句
		if strings.HasSuffix(trimmed, ";") {
			s := strings.TrimSpace(current.String())
			if s != "" {
				stmts = append(stmts, s)
			}
			current.Reset()
		}
	}

	if current.Len() > 0 {
		if s := strings.TrimSpace(current.String()); s != "" {
			stmts = append(stmts, s)
		}
	}

	return stmts
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
