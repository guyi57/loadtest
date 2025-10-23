package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

var cronJob = cron.New()

func restoreTasksFromDB() {
	var tasks []ScheduledTask
	if err := db.Find(&tasks).Error; err != nil {
		log.Println("恢复任务失败:", err)
		return
	}

	for i := range tasks {
		task := &tasks[i]
		entryID, err := cronJob.AddFunc(task.CronExpr, func() {
			broadcastLog(fmt.Sprintf("🕐 定时任务开始执行 %s", time.Now().Format("2006-01-02 15:04:05")))
			executeScheduledTask(task)
		})

		if err != nil {
			log.Printf("恢复任务 [%s] 失败: %v\n", task.TaskID, err)
			continue
		}

		task.EntryID = entryID
		log.Printf("✅ 恢复定时任务 [%s]: %s %s\n", task.TaskID, task.Method, task.URL)
	}

	if len(tasks) > 0 {
		broadcastLog(fmt.Sprintf("📦 已从数据库恢复 %d 个定时任务", len(tasks)))
	}
}

func executeScheduledTask(task *ScheduledTask) {
	var headers map[string]string
	if task.Headers != "" {
		json.Unmarshal([]byte(task.Headers), &headers)
	}

	req := LoadTestRequest{
		TotalRequests: task.TotalRequests,
		Concurrency:   task.Concurrency,
		Method:        task.Method,
		URL:           task.URL,
		Headers:       headers,
		Body:          task.Body,
	}

	executeLoadTestWithLog(req, "scheduled", task.TaskID)
}

func getTasksHandler(c *gin.Context) {
	var tasks []ScheduledTask
	if err := db.Find(&tasks).Error; err != nil {
		c.JSON(500, gin.H{"error": "查询任务失败"})
		return
	}

	c.JSON(200, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

func deleteTaskHandler(c *gin.Context) {
	taskID := c.Param("id")

	var task ScheduledTask
	if err := db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(404, gin.H{"error": "任务不存在"})
		return
	}

	if task.EntryID != 0 {
		cronJob.Remove(task.EntryID)
	}

	if err := db.Delete(&task).Error; err != nil {
		c.JSON(500, gin.H{"error": "删除任务失败"})
		return
	}

	broadcastLog(fmt.Sprintf("🗑️ 定时任务已取消 [%s]: %s %s", taskID, task.Method, task.URL))
	c.JSON(200, gin.H{"message": "任务已取消", "task_id": taskID})
}
