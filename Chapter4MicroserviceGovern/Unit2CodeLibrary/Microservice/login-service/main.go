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

/* ---------- DTO ---------- */

type loginRequest struct{ Username, Password string }
type registerRequest struct{ Username, Password string }
type loginResponse struct {
	Success   bool   `json:"success"`
	AuthToken string `json:"authToken,omitempty"`
	ID        string `json:"id,omitempty"`
}

/* ---------- token ---------- */

func generateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func generateAuthToken() (string, error) { return generateRandomToken(32) }

/* ---------- Cloud-Native 访问日志 ---------- */

func accessLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		lat := fmt.Sprintf("%.3fs", time.Since(start).Seconds())
		req := c.Request

		httpReq := zap.Object("httpRequest", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
			enc.AddString("requestMethod", req.Method)
			enc.AddString("requestUrl", req.URL.Path+"?"+req.URL.RawQuery)
			enc.AddInt("status", c.Writer.Status())
			enc.AddString("latency", lat)
			enc.AddString("remoteIp", c.ClientIP())
			enc.AddString("userAgent", req.UserAgent())
			enc.AddString("protocol", req.Proto)
			return nil
		}))

		ent := logger.With(httpReq, zap.String("service", "login-service"))

		if tid := c.GetHeader("Trace-ID"); tid != "" {
			ent = ent.With(zap.String("trace", tid))
		}
		if sid := c.GetHeader("Span-ID"); sid != "" {
			ent = ent.With(zap.String("spanId", sid))
		}
		if uid, ok := c.Get("userID"); ok {
			ent = ent.With(zap.Object("labels", zapcore.ObjectMarshalerFunc(
				func(enc zapcore.ObjectEncoder) error {
					enc.AddString("userID", fmt.Sprint(uid))
					return nil
				})))
		}
		ent.Info("") // severity=INFO
	}
}

/* ---------- Handlers ---------- */

func loginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid JSON"})
		return
	}
	var user User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(401, gin.H{"error": "user not found"})
		} else {
			c.JSON(500, gin.H{"error": "db error"})
		}
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil &&
		user.Password != req.Password {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	if user.AuthToken == "" {
		tok, _ := generateAuthToken()
		user.AuthToken = tok
		_ = db.Save(&user).Error
	}
	writeAuthCookies(c, user.AuthToken, user.ID)
	c.Set("userID", user.ID)
	c.JSON(200, loginResponse{true, user.AuthToken, user.ID})
}

func registerHandler(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid JSON"})
		return
	}
	var exist User
	if err := db.Where("Username = ?", req.Username).First(&exist).Error; err == nil {
		c.JSON(409, gin.H{"error": "username exists"})
		return
	} else if !gorm.IsRecordNotFoundError(err) {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	id, err := getNextUserID()
	if err != nil {
		c.JSON(500, gin.H{"error": "id error"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := User{ID: id, Username: req.Username, Password: string(hash)}
	_ = db.Create(&user).Error
	writeAuthCookies(c, user.AuthToken, user.ID)
	c.Set("userID", user.ID)
	c.JSON(201, loginResponse{true, user.AuthToken, user.ID})
}

func userHandler(c *gin.Context) {
	tok := c.GetHeader("Authorization")
	uid := c.GetHeader("X-User-ID")
	if tok == "" {
		tok, _ = c.Cookie("AuthToken")
	}
	if uid == "" {
		uid, _ = c.Cookie("X-User-ID")
	}
	if tok == "" || uid == "" {
		c.JSON(401, gin.H{"error": "missing auth"})
		return
	}
	var u User
	if err := db.Where("AuthToken=? AND ID=?", tok, uid).First(&u).Error; err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	c.Set("userID", u.ID)
	c.JSON(200, u)
}

/* ---------- util ---------- */

func writeAuthCookies(c *gin.Context, token, id string) {
	age := 7 * 24 * 3600
	c.SetCookie("AuthToken", token, age, "/", "", false, true)
	c.SetCookie("X-User-ID", id, age, "/", "", false, true)
}

func newRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(accessLogger(), gin.Recovery())
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
	return r
}

/* ---------- main ---------- */

func main() {
	initNacos()
	initDatabase()
	defer closeDatabase()
	defer logger.Sync()

	srv := &http.Server{Addr: ":8083", Handler: newRouter()}

	go func() {
		logger.Info("login-service listening", zap.String("service", "login-service"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen", zap.Error(err))
		}
	}()

	ip, _ := getHostIP()
	if err := registerService(NamingClient, "login-service", ip, 8083); err != nil {
		logger.Fatal("register", zap.Error(err))
	}
	defer deregisterLoginService()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	logger.Info("shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("server exit")
}
