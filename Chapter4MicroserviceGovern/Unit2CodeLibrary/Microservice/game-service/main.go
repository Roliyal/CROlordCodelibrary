// main.go
package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// 全局 logger
var zapLog *zap.SugaredLogger

func initLogger() {
	cfg := zap.NewProductionConfig()
	// 1) 统一时间格式
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 2) 不要 caller 全路径，只保留文件名:行号
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	// 3) 用环境变量控制最低级别：INFO | WARN | ERROR | DEBUG
	level := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(level) {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	// 4) 打开采样，压缩重复日志
	cfg.Sampling = &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}
	logger, _ := cfg.Build()
	zapLog = logger.Sugar()
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

// ZapRequestLogger 把 Gin 访问日志写到 zap
func ZapRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		traceID := c.GetHeader("traceparent")

		zapLog.Infow("HTTP",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"latency", latency.String(),
			"size", c.Writer.Size(),
			"ip", c.ClientIP(),
			"trace_id", traceID,
		)
	}
}

func main() {
	initLogger()
	defer zapLog.Sync()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(ZapRequestLogger(), gin.Recovery())

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://micro.roliyal.com")
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
	zapLog.Infof("User guessed number: %d", req.Number)

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
