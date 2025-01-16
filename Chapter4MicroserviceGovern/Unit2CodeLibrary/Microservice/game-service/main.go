// main.go

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// 定义请求和响应结构体
type guessRequest struct {
	Number int `json:"number"`
}

type guessResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Attempts int    `json:"attempts"`
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Success   bool   `json:"success"`
	AuthToken string `json:"authToken"`
	ID        string `json:"id"` // 使用字符串类型
}

// respondWithError 统一错误响应
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

func main() {
	// 初始化日志目录
	logDir := "/app/log"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, 0777)
	}

	// 初始化 Nacos
	initNacos()

	// 注册 game-service 到 Nacos
	err := registerService(NamingClient, "game-service", "127.0.0.1", 8084)
	if err != nil {
		fmt.Printf("Error registering game service instance: %v\n", err)
		os.Exit(1)
	}

	// 订阅 login-service 的变化
	subscribeLoginService()

	// 获取并初始化数据库配置
	dbConfig, err := getDatabaseConfigFromNacos()
	if err != nil {
		panic("failed to get database configuration from Nacos")
	}
	initDatabase(dbConfig) // Initialize the database with the configuration from Nacos
	defer closeDatabase()

	// 设置 HTTP 路由
	mux := http.NewServeMux()
	mux.HandleFunc("/game", guessHandler)
	mux.HandleFunc("/health", healthCheckHandler)

	// 应用 CORS 中间件
	handler := corsMiddleware(mux)

	// 启动 HTTP 服务器
	fmt.Println("Starting server on port 8084")
	go func() {
		log.Fatal(http.ListenAndServe(":8084", handler))
	}()

	// 处理优雅关闭
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// 注销 game-service
	deregisterGameService()
}

// healthCheckHandler 健康检查处理器
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// corsMiddleware CORS 中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置允许的来源
		w.Header().Set("Access-Control-Allow-Origin", "http://micro.roliyal.com")

		// 设置允许的请求头，包括自定义头 'X-User-ID' 和 'Authorization'
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID, Authorization")

		// 设置允许的HTTP方法
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// 允许携带凭证
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// 处理预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 调用下一个处理器
		next.ServeHTTP(w, r)
	})
}

// guessHandler 处理猜数字请求
func guessHandler(w http.ResponseWriter, r *http.Request) {
	// 输出请求头，确保 X-User-ID 和 Authorization 被接收到
	log.Printf("Received headers: %v", r.Header)

	userIdStr := r.Header.Get("X-User-ID")     // 读取 X-User-ID 请求头
	authToken := r.Header.Get("Authorization") // 读取 Authorization 请求头

	if userIdStr == "" {
		log.Println("Error: Missing X-User-ID header")
		respondWithError(w, http.StatusBadRequest, "Missing X-User-ID header")
		return
	}

	// 使用 userIdStr 和 authToken 从 login-service 获取用户信息
	user, err := getUserFromUserID(userIdStr, authToken)
	if err != nil {
		log.Printf("Error getting user: %v\n", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 读取请求体
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	// 解析请求体
	var req guessRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// 获取或创建游戏记录
	game, err := getOrCreateGame(&user)
	if err != nil {
		log.Println("Error getting or creating game:", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// 处理猜测逻辑
	var res guessResponse
	if req.Number == game.TargetNumber {
		res.Success = true
		res.Message = "Congratulations! You guessed the correct number."
		res.Attempts = game.Attempts
		game.CorrectGuesses++
		if err := db.Save(game).Error; err != nil {
			log.Printf("Error updating game: %v", err)
		}
	} else {
		res.Success = false
		if req.Number < game.TargetNumber {
			res.Message = "The number is too low."
		} else {
			res.Message = "The number is too high."
		}
		incrementAttempts(game)
		res.Attempts = game.Attempts
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
