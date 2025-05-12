package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	_ "google.golang.org/protobuf/proto"
	_ "google.golang.org/protobuf/runtime/protoimpl"
)

/* ----------------- DTO ----------------- */

type (
	loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	registerRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	loginResponse struct {
		Success   bool   `json:"success"`
		AuthToken string `json:"authToken,omitempty"`
		ID        string `json:"id,omitempty"`
	}
)

/* ----------------- token helpers ----------------- */

func generateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func generateAuthToken() (string, error) { return generateRandomToken(32) }

/* ----------------- handlers ----------------- */

// 记录请求的详细信息
func logRequestDetails(c *gin.Context) {
	method := c.Request.Method
	path := c.Request.URL.Path
	queryParams := c.Request.URL.Query().Encode() // 获取请求的查询字符串

	// 记录请求的详细信息
	logger.Info("Request Details",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("query", queryParams))

	// 记录请求体
	if method == "POST" {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			logger.Info("Request Body", zap.Any("body", body))
		}
	}
}

// 登录处理
func loginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	var user User
	if err := db.Select("ID, Username, Password, AuthToken, Wins, Attempts").
		Where("username = ?", req.Username).First(&user).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			logger.Warn("User not found", zap.String("username", req.Username))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		} else {
			logger.Error("DB error", zap.String("username", req.Username), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		}
		return
	}

	passOK := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) == nil ||
		user.Password == req.Password
	if !passOK {
		logger.Warn("Invalid credentials", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// 登录成功，生成 token
	token := user.AuthToken
	if token == "" {
		var err error
		token, err = generateAuthToken()
		if err != nil {
			logger.Error("Token generation error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
			return
		}
		user.AuthToken = token
		_ = db.Save(&user).Error
	}

	// 记录成功的登录日志
	logger.Info("User logged in",
		zap.String("username", req.Username),
		zap.String("authToken", token),
		zap.String("userID", user.ID))

	// 设置 cookies
	writeAuthCookies(c, token, user.ID)
	c.JSON(http.StatusOK, loginResponse{Success: true, AuthToken: token, ID: user.ID})
}

// 注册处理
func registerHandler(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	var exist User
	if err := db.Where("Username = ?", req.Username).First(&exist).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username exists"})
		return
	} else if !gorm.IsRecordNotFoundError(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	// 获取下一个用户ID
	nextID, err := getNextUserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "id error"})
		return
	}

	// 加密密码并保存用户
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := User{ID: nextID, Username: req.Username, Password: string(hash)}
	if err = db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	// 设置 cookies 并返回
	writeAuthCookies(c, user.AuthToken, user.ID)
	c.JSON(http.StatusCreated, loginResponse{Success: true, AuthToken: user.AuthToken, ID: user.ID})
}

// 获取用户信息
func userHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	userID := c.GetHeader("X-User-ID")
	if authToken == "" {
		authToken, _ = c.Cookie("AuthToken")
	}
	if userID == "" {
		userID, _ = c.Cookie("X-User-ID")
	}

	if authToken == "" || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth"})
		return
	}

	var user User
	if err := db.Where("AuthToken = ? AND ID = ?", authToken, userID).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		}
		return
	}
	c.JSON(http.StatusOK, user)
}

/* ----------------- cookie util ----------------- */

// 写入认证cookie
func writeAuthCookies(c *gin.Context, token, id string) {
	age := 7 * 24 * 3600
	c.SetCookie("AuthToken", token, age, "/", "", false, true)
	c.SetCookie("X-User-ID", id, age, "/", "", false, true)
}

func main() {
	/* ------- 初始化 ------- */
	initNacos()
	initDatabase()
	defer closeDatabase()
	defer logger.Sync()

	/* ------- Gin & zap middleware ------- */
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://micro.roliyal.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-User-ID"},
		AllowCredentials: true,
	}))
	r.POST("/login", loginHandler)
	r.POST("/register", registerHandler)
	r.GET("/user", userHandler)
	r.GET("/health", func(c *gin.Context) { c.String(200, "ok") })

	// 启动 HTTP 服务
	srv := &http.Server{Addr: ":8083", Handler: r}

	/* ------- HTTP serve ------- */
	go func() {
		logger.Info("login-service listening", zap.String("addr", ":8083"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen err", zap.Error(err))
		}
	}()

	/* ------- 注册到 Nacos ------- */
	hostIP, err := getHostIP()
	if err != nil {
		logger.Fatal("get host ip", zap.Error(err))
	}
	if err = registerService(NamingClient, "login-service", hostIP, 8083); err != nil {
		logger.Fatal("register service", zap.Error(err))
	}
	defer deregisterLoginService()

	/* ------- 优雅关机 ------- */
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logger.Info("termination signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("http shutdown", zap.Error(err))
	}
	logger.Info("server exited gracefully")
}
