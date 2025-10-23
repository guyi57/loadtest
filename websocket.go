package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级错误:", err)
		return
	}

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	broadcastLog("🔌 新客户端已连接")

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func broadcastLog(message string) {
	log.Println(message)

	clientsMu.Lock()
	defer clientsMu.Unlock()

	timestamp := time.Now().Format("15:04:05")
	logMessage := map[string]string{
		"timestamp": timestamp,
		"message":   message,
	}

	data, _ := json.Marshal(logMessage)

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
