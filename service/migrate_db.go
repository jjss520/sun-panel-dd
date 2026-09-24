package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbPath := "/data/database/database.db"
	
	// 如果命令行参数提供了路径,使用参数
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}
	
	fmt.Printf("数据库路径: %s\n", dbPath)
	
	// 打开数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("无法打开数据库:", err)
	}
	
	fmt.Println("✓ 数据库连接成功")
	
	// 读取SQL文件
	sqlContent, err := ioutil.ReadFile("migrate_add_pages_sqlite.sql")
	if err != nil {
		log.Fatal("无法读取SQL文件:", err)
	}
	
	// 分割SQL语句(按分号分割)
	statements := strings.Split(string(sqlContent), ";")
	
	fmt.Printf("找到 %d 条SQL语句\n", len(statements))
	
	// 执行每条SQL
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		
		fmt.Printf("[%d/{}] 执行: %s...\n", i+1, len(statements), strings.Split(stmt, "\n")[0][:min(50, len(strings.Split(stmt, "\n")[0]))])
		
		if err := db.Exec(stmt).Error; err != nil {
			fmt.Printf("  ✗ 错误: %v\n", err)
		} else {
			fmt.Println("  ✓ 成功")
		}
	}
	
	fmt.Println("\n✓ 数据库迁移完成!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
