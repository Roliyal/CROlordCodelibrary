package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// 定义请求数据结构
type MCPRequest struct {
	ContextData string `json:"context_data"`
}

// 定义响应数据结构
type MCPResponse struct {
	ResponseData string `json:"response_data"`
}

// 存储有效的 API Key（示例中用硬编码，实际中应存储在安全地方）
var validAPIKey = "12345-api-key"

// API Key 验证函数
func validateAPIKey(r *http.Request) bool {
	// 获取请求头中的 Authorization 字段
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	// 检查 Authorization 头部的格式
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || parts[0] != "APIKey" {
		return false
	}

	// 验证 API Key
	apiKey := parts[1]
	return apiKey == validAPIKey
}

// MCP 服务处理函数
func mcpHandler(w http.ResponseWriter, r *http.Request) {
	// 验证 API Key
	if !validateAPIKey(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 只接受 POST 请求
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求的 JSON 数据
	var req MCPRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	// 这里简单地返回接收到的上下文数据
	response := MCPResponse{
		ResponseData: fmt.Sprintf("Received context: %s", req.ContextData),
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 返回响应的 JSON 数据
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 设置路由和处理器
	http.HandleFunc("/mcp", mcpHandler)

	// 启动 HTTP 服务器
	port := ":8080"
	fmt.Println("Starting MCP server on port", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
