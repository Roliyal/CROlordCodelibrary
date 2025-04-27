// main.go — 完整源代码（Go 1.22 可直接编译）
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/bcrypt"
)

/* ------------------------------------------------------------------
                               常量
-------------------------------------------------------------------*/

const (
	serviceName = "login-service"  // 顶层字段 service
	projectID   = "micro-go-login" // trace 字段拼接用
)

/* ------------------------------------------------------------------
                               DTO
-------------------------------------------------------------------*/

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type loginResponse struct {
	Success   bool   `json:"success"`
	AuthToken string `json:"authToken,omitempty"`
	ID        string `json:"id,omitempty"`
}

/* ------------------------------------------------------------------
                         Token / 辅助函数
-------------------------------------------------------------------*/

func generateAuthToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeAuthCookies(c *gin.Context, token, id string) {
	age := 7 * 24 * 3600
	c.SetCookie("AuthToken", token, age, "/", "", false, true)
	c.SetCookie("X-User-ID", id, age, "/", "", false, true)
}

/* ------------------------------------------------------------------
                        自定义 zap 访问日志
-------------------------------------------------------------------*/

// 返回一个 Gin middleware，把访问日志写成 CNCF / Google Cloud JSON
func zapLogger(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // 先走后续 handler

		latency := fmt.Sprintf("%.3fs", time.Since(start).Seconds())
		url := c.Request.URL.Path
		if q := c.Request.URL.RawQuery; q != "" {
			url += "?" + q
		}

		traceID := c.GetHeader("Trace-ID")
		spanID := c.GetHeader("Span-ID")

		fields := []zap.Field{
			zap.String("service", serviceName),
			zap.Object("httpRequest", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
				enc.AddString("requestMethod", c.Request.Method)
				enc.AddString("requestUrl", url)
				enc.AddInt("status", c.Writer.Status())
				enc.AddString("latency", latency)
				enc.AddString("remoteIp", c.ClientIP())
				enc.AddString("userAgent", c.Request.UserAgent())
				enc.AddString("protocol", c.Request.Proto)
				return nil
			})),
		}
		if traceID != "" {
			fields = append(fields, zap.String("trace",
				"projects/"+projectID+"/traces/"+traceID))
		}
		if spanID != "" {
			fields = append(fields, zap.String("spanId", spanID))
		}

		l.Info("", fields...)
	}
}

/* ------------------------------------------------------------------
                          HTTP Handlers
-------------------------------------------------------------------*/

func loginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	var user User
	if err := db.Select("ID, Username, Password, AuthToken").
		Where("username = ?", req.Username).First(&user).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		}
		return
	}

	passOK := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) == nil ||
		user.Password == req.Password
	if !passOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token := user.AuthToken
	if token == "" {
		var err error
		token, err = generateAuthToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
			return
		}
		user.AuthToken = token
		_ = db.Save(&user).Error
	}

	writeAuthCookies(c, token, user.ID)
	c.JSON(http.StatusOK, loginResponse{Success: true, AuthToken: token, ID: user.ID})
}

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

	nextID, err := getNextUserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "id error"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := User{ID: nextID, Username: req.Username, Password: string(hash)}
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	writeAuthCookies(c, user.AuthToken, user.ID)
	c.JSON(http.StatusCreated, loginResponse{Success: true, AuthToken: user.AuthToken, ID: user.ID})
}

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
	if err := db.Where("AuthToken = ? AND ID = ?", authToken, userID).
		First(&user).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		}
		return
	}
	c.JSON(http.StatusOK, user)
}

/* ------------------------------------------------------------------
                               main
-------------------------------------------------------------------*/

func main() {
	/* 初始化 Nacos & DB */
	initNacos()
	initDatabase()
	defer closeDatabase()
	defer logger.Sync()

	/* Gin + 中间件 */
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(zapLogger(logger)) // 访问日志
	r.Use(gin.Recovery())    // panic 保护
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://micro.roliyal.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-User-ID"},
		AllowCredentials: true,
	}))

	/* 路由 */
	r.POST("/login", loginHandler)
	r.POST("/register", registerHandler)
	r.GET("/user", userHandler)
	r.GET("/health", func(c *gin.Context) { c.String(200, "ok") })

	/* HTTP server */
	srv := &http.Server{Addr: ":8083", Handler: r}
	go func() {
		logger.Info("server listening", zap.String("addr", ":8083"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen", zap.Error(err))
		}
	}()

	/* Nacos 注册 */
	hostIP, err := getHostIP()
	if err != nil {
		logger.Fatal("get host ip", zap.Error(err))
	}
	if err = registerService(NamingClient, serviceName, hostIP, 8083); err != nil {
		logger.Fatal("register service", zap.Error(err))
	}
	defer deregisterLoginService()

	/* 优雅关机 */
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
