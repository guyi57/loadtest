package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func executeLoadTest(req LoadTestRequest) {
	executeLoadTestWithLog(req, "immediate", "")
}

func executeLoadTestWithLog(req LoadTestRequest, executeType, taskID string) {
	startTime := time.Now()
	broadcastLog(fmt.Sprintf("🚀 开始压力测试: 总请求数 %d，并发数 %d", req.TotalRequests, req.Concurrency))
	broadcastLog(fmt.Sprintf("📍 目标地址: %s %s", req.Method, req.URL))

	var (
		successCount    int32
		failureCount    int32
		wg              sync.WaitGroup
		semaphore       = make(chan struct{}, req.Concurrency)
		responseTimes   []time.Duration
		responseTimesMu sync.Mutex
	)

	for i := 0; i < req.TotalRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(index int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			reqStart := time.Now()
			err := makeRequest(req.Method, req.URL, req.Headers, req.Body)
			duration := time.Since(reqStart)

			responseTimesMu.Lock()
			responseTimes = append(responseTimes, duration)
			responseTimesMu.Unlock()

			if err != nil {
				atomic.AddInt32(&failureCount, 1)
				broadcastLog(fmt.Sprintf("❌ 请求 #%d 失败: %v (%.2fms)", index+1, err, float64(duration.Microseconds())/1000))
			} else {
				atomic.AddInt32(&successCount, 1)
				if (index+1)%10 == 0 || index == 0 {
					broadcastLog(fmt.Sprintf("✅ 请求 #%d 成功 (%.2fms)", index+1, float64(duration.Microseconds())/1000))
				}
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	result := calculateResults(int(successCount), int(failureCount), totalDuration, responseTimes)

	broadcastLog(strings.Repeat("=", 60))
	broadcastLog("📊 测试结果:")
	broadcastLog(fmt.Sprintf("   总请求数: %d", result.TotalRequests))
	broadcastLog(fmt.Sprintf("   成功: %d (%.2f%%)", result.SuccessCount, float64(result.SuccessCount)/float64(result.TotalRequests)*100))
	broadcastLog(fmt.Sprintf("   失败: %d (%.2f%%)", result.FailureCount, float64(result.FailureCount)/float64(result.TotalRequests)*100))
	broadcastLog(fmt.Sprintf("   总耗时: %v", result.TotalDuration))
	broadcastLog(fmt.Sprintf("   平均响应时间: %v", result.AvgResponseTime))
	broadcastLog(fmt.Sprintf("   最小响应时间: %v", result.MinResponseTime))
	broadcastLog(fmt.Sprintf("   最大响应时间: %v", result.MaxResponseTime))
	broadcastLog(fmt.Sprintf("   每秒请求数: %.2f", result.RequestsPerSec))
	broadcastLog(strings.Repeat("=", 60))

	testLog := TestLog{
		URL:             req.URL,
		Method:          req.Method,
		TotalRequests:   req.TotalRequests,
		Concurrency:     req.Concurrency,
		SuccessCount:    result.SuccessCount,
		FailureCount:    result.FailureCount,
		TotalDuration:   result.TotalDuration.Milliseconds(),
		AvgResponseTime: result.AvgResponseTime.Milliseconds(),
		MinResponseTime: result.MinResponseTime.Milliseconds(),
		MaxResponseTime: result.MaxResponseTime.Milliseconds(),
		RequestsPerSec:  result.RequestsPerSec,
		ExecuteType:     executeType,
		TaskID:          taskID,
		CreatedAt:       time.Now(),
	}

	if err := db.Create(&testLog).Error; err != nil {
		log.Printf("保存测试日志失败: %v\n", err)
	} else {
		broadcastLog(fmt.Sprintf("💾 测试日志已保存 [ID: %d]", testLog.ID))
	}
}

func makeRequest(method, url string, headers map[string]string, body string) error {
	var req *http.Request
	var err error

	if method == "POST" {
		req, err = http.NewRequest(method, url, bytes.NewBufferString(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil
}

func calculateResults(success, failure int, totalDuration time.Duration, responseTimes []time.Duration) TestResult {
	total := success + failure
	result := TestResult{
		TotalRequests: total,
		SuccessCount:  success,
		FailureCount:  failure,
		TotalDuration: totalDuration,
	}

	if len(responseTimes) > 0 {
		var sum time.Duration
		min := responseTimes[0]
		max := responseTimes[0]

		for _, rt := range responseTimes {
			sum += rt
			if rt < min {
				min = rt
			}
			if rt > max {
				max = rt
			}
		}

		result.AvgResponseTime = sum / time.Duration(len(responseTimes))
		result.MinResponseTime = min
		result.MaxResponseTime = max
	}

	if totalDuration.Seconds() > 0 {
		result.RequestsPerSec = float64(total) / totalDuration.Seconds()
	}

	return result
}
