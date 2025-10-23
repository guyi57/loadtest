package main

import (
	"time"

	"github.com/robfig/cron/v3"
)

type LoadTestRequest struct {
	TotalRequests int               `json:"total_requests" binding:"required,min=1"`
	Concurrency   int               `json:"concurrency" binding:"required,min=1"`
	Method        string            `json:"method" binding:"required,oneof=GET POST"`
	URL           string            `json:"url" binding:"required,url"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
	ExecuteType   string            `json:"execute_type" binding:"required,oneof=immediate scheduled"`
	CronExpr      string            `json:"cron_expr"`
}

type TestResult struct {
	TotalRequests   int           `json:"total_requests"`
	SuccessCount    int           `json:"success_count"`
	FailureCount    int           `json:"failure_count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	MinResponseTime time.Duration `json:"min_response_time"`
	MaxResponseTime time.Duration `json:"max_response_time"`
	RequestsPerSec  float64       `json:"requests_per_sec"`
}

type ScheduledTask struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	TaskID         string        `gorm:"uniqueIndex;size:50" json:"task_id"`
	CronExpr       string        `gorm:"size:100" json:"cron_expr"`
	URL            string        `gorm:"size:500" json:"url"`
	Method         string        `gorm:"size:10" json:"method"`
	TotalRequests  int           `json:"total_requests"`
	Concurrency    int           `json:"concurrency"`
	Headers        string        `gorm:"type:text" json:"headers"`
	Body           string        `gorm:"type:text" json:"body"`
	CreatedAt      time.Time     `json:"created_at"`
	EntryID        cron.EntryID  `gorm:"-" json:"-"`
}

type TestLog struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	URL             string    `gorm:"size:500" json:"url"`
	Method          string    `gorm:"size:10" json:"method"`
	TotalRequests   int       `json:"total_requests"`
	Concurrency     int       `json:"concurrency"`
	SuccessCount    int       `json:"success_count"`
	FailureCount    int       `json:"failure_count"`
	TotalDuration   int64     `json:"total_duration"`
	AvgResponseTime int64     `json:"avg_response_time"`
	MinResponseTime int64     `json:"min_response_time"`
	MaxResponseTime int64     `json:"max_response_time"`
	RequestsPerSec  float64   `json:"requests_per_sec"`
	ExecuteType     string    `gorm:"size:20" json:"execute_type"`
	TaskID          string    `gorm:"size:50" json:"task_id"`
	CreatedAt       time.Time `json:"created_at"`
}
