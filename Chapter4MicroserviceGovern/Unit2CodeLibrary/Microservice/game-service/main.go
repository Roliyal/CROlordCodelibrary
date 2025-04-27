// main.go
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// 全局 logger
var zapLog *zap.SugaredLogger

// 初始化 logger
func initLogger() {
	z, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("cannot initialize zap logger: %v", err))
	}
	zapLog = z.Sugar()
}

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
	initLogger()
	defer zapLog.Sync()

	// 创建 Gin 引擎
	r := gin.Default()

	// CORS 配置
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://micro.roliyal.com") // 前端地址
		c.Header("Access-Control-Allow-Credentials", "true")                // 允许携带 Cookies
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
		zapLog.Fatalf("Error registering game service instance: %v", err)
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
			zapLog.Fatalf("Error starting server: %v", err)
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
	zapLog.Infof("Received headers: %v", c.Request.Header)

	// 打印所有 cookies
	cookies := c.Request.Cookies()
	zapLog.Infof("Received cookies: %v", cookies)

	userIdStr, err := c.Cookie("X-User-ID")
	if err != nil || userIdStr == "" {
		userIdStr = c.GetHeader("X-User-ID")
	}
	if userIdStr == "" {
		zapLog.Error("Missing X-User-ID from Cookie or Header")
		respondWithError(c, http.StatusBadRequest, "Missing X-User-ID")
		return
	}
	zapLog.Infof("Got X-User-ID: %s", userIdStr)

	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		zapLog.Warn("Missing Authorization header")
		respondWithError(c, http.StatusUnauthorized, "Missing Authorization token")
		return
	}
	zapLog.Infof("Got Authorization: %s", authToken)

	user, err := getUserFromUserID(userIdStr, authToken)
	if err != nil {
		zapLog.Errorf("Error getting user from login-service: %v", err)
		respondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	//  读取 JSON 请求体
	var req guessRequest
	if err := c.BindJSON(&req); err != nil {
		zapLog.Errorf("Error decoding request body:", err)
		respondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	zapLog.Infof("📥 User guessed number: %d", req.Number)

	//  获取或创建游戏记录
	game, err := getOrCreateGame(&user)
	if err != nil {
		zapLog.Errorf("Error getting or creating game:", err)
		respondWithError(c, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	//  猜数字逻辑
	var res guessResponse
	if req.Number == game.TargetNumber {
		res.Success = true
		res.Message = " Congratulations! You guessed the correct number. this is gray"
		res.Attempts = game.Attempts
		game.CorrectGuesses++
		if err := db.Save(game).Error; err != nil {
			zapLog.Errorf("Error updating game: %v", err)
		}
	} else {
		res.Success = false
		if req.Number < game.TargetNumber {
			res.Message = " Too low. Try again!this is gray"
		} else {
			res.Message = " Too high. Try again!this is gray"
		}
		incrementAttempts(game)
		res.Attempts = game.Attempts
	}

	//  返回 JSON 响应
	c.JSON(http.StatusOK, res)
}
