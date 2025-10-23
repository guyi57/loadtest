package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	// 初始化数据库
	if err := initDatabase(); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// 启动 cron 并恢复已保存的任务
	cronJob.Start()
	defer cronJob.Stop()
	restoreTasksFromDB()

	// 查找可用端口
	startPort := 8080
	port := findAvailablePort(startPort)

	if port != startPort {
		log.Printf("⚠️  端口 %d 被占用，自动切换到端口 %d\n", startPort, port)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 使用嵌入的模板文件
	templ := template.Must(template.New("").ParseFS(templatesFS, "templates/*.html"))
	r.SetHTMLTemplate(templ)

	// 路由
	r.GET("/", indexHandler)
	r.POST("/api/test", loadTestHandler)
	r.GET("/api/tasks", getTasksHandler)
	r.DELETE("/api/tasks/:id", deleteTaskHandler)
	r.GET("/api/logs", getLogsHandler)
	r.DELETE("/api/logs/:id", deleteLogHandler)
	r.GET("/ws", wsHandler)

	// 服务器地址
	serverURL := fmt.Sprintf("http://localhost:%d", port)
	log.Printf("🚀 服务器启动成功: %s\n", serverURL)
	log.Println("📦 HTML 资源已嵌入，无需 templates 目录")
	log.Println("💾 数据库文件: loadtest.db")

	// 延迟打开浏览器
	go func() {
		time.Sleep(500 * time.Millisecond)
		log.Println("🌐 正在打开浏览器...")
		openBrowser(serverURL)
	}()

	// 启动服务器
	addr := fmt.Sprintf(":%d", port)
	if err := r.Run(addr); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
