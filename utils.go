package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"
)

// 查找可用端口
func findAvailablePort(startPort int) int {
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return startPort
}

// 检查端口是否可用
func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// 自动打开浏览器
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("无法自动打开浏览器: %v\n", err)
	}
}
