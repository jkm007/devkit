package main

import (
	"log"

	"backend-server/config"
	"backend-server/migrations"
	"backend-server/pkg/database"
)

func main() {
	// 加载配置
	cfg, err := config.Load("./config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	if err := database.InitMySQL(cfg.MySQL); err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer database.CloseMySQL()

	// 运行迁移
	log.Println("开始运行数据库迁移...")
	if err := migrations.Run(database.GetMySQL()); err != nil {
		log.Fatalf("运行迁移失败: %v", err)
	}

	log.Println("数据库迁移完成!")
}
