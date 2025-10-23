package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func loadTestHandler(c *gin.Context) {
	var req LoadTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Concurrency > req.TotalRequests {
		req.Concurrency = req.TotalRequests
	}

	if req.ExecuteType == "scheduled" {
		if req.CronExpr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "定时任务需要提供 cron 表达式"})
			return
		}

		headersJSON, _ := json.Marshal(req.Headers)

		task := &ScheduledTask{
			CronExpr:      req.CronExpr,
			URL:           req.URL,
			Method:        req.Method,
			TotalRequests: req.TotalRequests,
			Concurrency:   req.Concurrency,
			Headers:       string(headersJSON),
			Body:          req.Body,
			CreatedAt:     time.Now(),
		}

		if err := db.Create(task).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存任务失败: " + err.Error()})
			return
		}

		task.TaskID = fmt.Sprintf("task_%d", task.ID)
		db.Save(task)

		entryID, err := cronJob.AddFunc(req.CronExpr, func() {
			broadcastLog(fmt.Sprintf("🕐 定时任务开始执行 %s", time.Now().Format("2006-01-02 15:04:05")))
			executeScheduledTask(task)
		})

		if err != nil {
			db.Delete(task)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无效的 cron 表达式: %v", err)})
			return
		}

		task.EntryID = entryID

		broadcastLog(fmt.Sprintf("✅ 定时任务创建成功 [%s]，cron 表达式: %s", task.TaskID, req.CronExpr))
		c.JSON(http.StatusOK, gin.H{"message": "定时任务创建成功", "task_id": task.TaskID, "cron": req.CronExpr})
		return
	}

	go executeLoadTestWithLog(req, "immediate", "")
	c.JSON(http.StatusOK, gin.H{"message": "压力测试已开始"})
}

func getLogsHandler(c *gin.Context) {
	var logs []TestLog

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "50")

	var pageInt, pageSizeInt int
	fmt.Sscanf(page, "%d", &pageInt)
	fmt.Sscanf(pageSize, "%d", &pageSizeInt)

	if pageInt < 1 {
		pageInt = 1
	}
	if pageSizeInt < 1 || pageSizeInt > 100 {
		pageSizeInt = 50
	}

	offset := (pageInt - 1) * pageSizeInt

	var total int64
	db.Model(&TestLog{}).Count(&total)

	if err := db.Order("created_at DESC").Limit(pageSizeInt).Offset(offset).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      pageInt,
		"page_size": pageSizeInt,
	})
}

func deleteLogHandler(c *gin.Context) {
	logID := c.Param("id")

	if err := db.Delete(&TestLog{}, logID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除日志失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "日志已删除"})
}
