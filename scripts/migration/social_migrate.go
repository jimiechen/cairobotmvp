// Social 域建表脚本执行器
//
// 用法：
//   cd scripts/migration && go run social_migrate.go
//   # 或指定自定义 SQL 文件：
//   cd scripts/migration && go run social_migrate.go ../custom.sql
//
// 默认读取同目录下的 004_social_domain_tables.sql（16 表 DDL）
//
// 环境变量（可选，有默认值）：
//   MYSQL_HOST     默认 127.0.0.1
//   MYSQL_PORT     默认 3306
//   MYSQL_USER     默认 root
//   MYSQL_PASSWORD 默认空
//   MYSQL_DATABASE 默认 go_biz
package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 确定 SQL 文件路径：命令行参数 > 默认路径
	sqlFile := "004_social_domain_tables.sql"
	if len(os.Args) > 1 {
		sqlFile = os.Args[1]
	}

	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "root")
	pass := getEnv("MYSQL_PASSWORD", "")
	dbName := getEnv("MYSQL_DATABASE", "go_biz")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		user, pass, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] 连接 MySQL 失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Ping MySQL 失败: %v\n", err)
		fmt.Fprintln(os.Stderr, "请确认 MySQL 服务已启动且连接参数正确")
		os.Exit(1)
	}

	fmt.Println("============================================")
	fmt.Println("  Social 域建表脚本执行器")
	fmt.Printf("  目标数据库: %s@%s:%s/%s\n", user, host, port, dbName)
	fmt.Printf("  SQL 文件:    %s\n", sqlFile)
	fmt.Println("============================================\n")

	// 确保目标数据库存在
	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] 创建数据库失败: %v\n", err)
		os.Exit(1)
	}
	if _, err := db.Exec(fmt.Sprintf("USE `%s`", dbName)); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] 切换数据库失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[OK] 已选择数据库: %s\n\n", dbName)

	// 读取并执行 SQL 文件
	content, err := os.ReadFile(sqlFile)
	if err != nil {
		// 尝试相对于脚本所在目录查找
		execDir, _ := os.Executable()
		baseDir := filepath.Dir(execDir)
		content, err = os.ReadFile(filepath.Join(baseDir, sqlFile))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] 读取 SQL 文件失败: %v\n", err)
			fmt.Fprintln(os.Stderr, "请确认文件路径正确，或使用绝对路径")
			os.Exit(1)
		}
	}

	fmt.Printf(">>> 读取 SQL 文件: %s (%d bytes)\n", sqlFile, len(content))

	statements := splitStatements(string(content))
	successCount := 0
	tableCount := 0

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			errMsg := strings.ToLower(err.Error())
			// 幂等操作警告可忽略
			if strings.Contains(errMsg, "already exists") ||
				strings.Contains(errMsg, "unknown table") {
				fmt.Printf("    [%02d] SKIP (幂等): %s\n", i+1, truncate(stmt, 60))
				continue
			}
			fmt.Fprintf(os.Stderr, "\n[ERROR] 语句[%02d] 执行失败:\n  SQL: %s\n  错误: %v\n", i+1, truncate(stmt, 120), err)
			os.Exit(1)
		}
		successCount++

		upper := strings.ToUpper(stmt)
		switch {
		case strings.Contains(upper, "CREATE TABLE"):
			tableCount++
			fmt.Printf("    [%02d] OK ✓ CREATE TABLE\n", i+1)
		case strings.Contains(upper, "DROP TABLE"):
			fmt.Printf("    [%02d] OK   DROP TABLE (清理旧表)\n", i+1)
		case strings.HasPrefix(upper, "USE"):
			// 静默
		case strings.HasPrefix(upper, "SET"):
			// 静默
		default:
			fmt.Printf("    [%02d] OK   %s\n", i+1, truncate(stmt, 50))
		}
	}

	fmt.Println("\n============================================")
	fmt.Printf("  执行完成: %d 条语句成功, %d 张表创建\n", successCount, tableCount)
	fmt.Println("============================================")

	// 验证：列出所有 Social 域表
	fmt.Println("\n--- 验证：查询 go_biz 库中的 Social 域表 ---")
	expectedTables := []string{
		"users", "groups", "topics", "group_members",
		"group_plans", "group_orders", "topic_read_records", "topic_comments",
		"topic_reactions", "topic_reply_likes", "group_admin_actions",
		"topic_audit_logs", "member_blocks", "member_stats", "group_stats", "topic_stats",
	}

	rows, err := db.Query(`SELECT table_name FROM information_schema.tables WHERE table_schema = ? ORDER BY table_name`, dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] 查询表列表失败: %v\n", err)
		return
	}
	defer rows.Close()

	var existingTables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			existingTables = append(existingTables, name)
		}
	}

	allOK := true
	for _, t := range expectedTables {
		found := false
		for _, e := range existingTables {
			if e == t {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("  [OK]   %s\n", t)
		} else {
			fmt.Printf("  [MISSING] %s ⚠️\n", t)
			allOK = false
		}
	}
	if allOK {
		fmt.Println("\n✅ 全部 16 张表验证通过")
	} else {
		fmt.Println("\n⚠️  部分表缺失，请检查外键依赖顺序或错误日志")
	}
}

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
