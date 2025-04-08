package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
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

// 统一错误响应
func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"success": false,
		"error":   message,
	})
}

func main() {
	// 创建 Gin 引擎
	r := gin.Default()

	// CORS 配置
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://micro.roliyal.com") // 前端地址
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-User-ID")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Next()
	})

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

	// 设置路由
	r.POST("/game", guessHandler)
	r.GET("/health", healthCheckHandler)

	// 启动 Gin HTTP 服务器
	go func() {
		if err := r.Run(":8084"); err != nil {
			log.Fatal("Error starting server: ", err)
		}
	}()

	// 处理优雅关闭
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// 注销 game-service
	deregisterGameService()
}

// healthCheckHandler 健康检查处理器
func healthCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

// guessHandler 处理猜数字请求
func guessHandler(c *gin.Context) {
	log.Printf("Received headers: %v", c.Request.Header)

	// 从 Cookie 中读取 "X-User-ID"
	userIdStr, err := c.Cookie("X-User-ID")
	if err != nil {
		log.Println("Error: Missing X-User-ID cookie")
		respondWithError(c, 400, "Missing X-User-ID cookie")
		return
	}

	// 读取 Authorization 头
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		log.Println("Warning: Missing Authorization header, but continuing anyway")
	}

	log.Printf("Got userIdStr from cookie: %s", userIdStr)
	log.Printf("Got authToken from header: %s", authToken)

	// 根据 userIdStr 和 authToken 获取用户信息
	user, err := getUserFromUserID(userIdStr, authToken)
	if err != nil {
		log.Printf("Error getting user: %v\n", err)
		respondWithError(c, 401, "Unauthorized")
		return
	}

	// 读取请求体
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		respondWithError(c, 400, "Invalid request body")
		return
	}
	defer c.Request.Body.Close()

	var req guessRequest
	if err := json.Unmarshal(body, &req); err != nil {
		respondWithError(c, 400, "Invalid JSON format")
		return
	}

	// 获取或创建游戏记录
	game, err := getOrCreateGame(&user)
	if err != nil {
		log.Println("Error getting or creating game:", err)
		respondWithError(c, 500, "Internal Server Error")
		return
	}

	// 进行猜数字逻辑
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

	c.JSON(200, res)
}
