// main.go
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

/* ----------- 用到的类型 ----------- */

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

/* ----------- token helpers ----------- */

func generateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func generateAuthToken() (string, error) { return generateRandomToken(32) }

/* ----------- handlers ----------- */

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

	user := User{
		ID:        nextID,
		Username:  req.Username,
		Password:  string(hash),
		AuthToken: "",
		Wins:      0,
		Attempts:  0,
	}
	if err = db.Create(&user).Error; err != nil {
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

/* ----------- cookie util ----------- */

func writeAuthCookies(c *gin.Context, token, id string) {
	age := 7 * 24 * 3600
	c.SetCookie("AuthToken", token, age, "/", "", false, true)
	c.SetCookie("X-User-ID", id, age, "/", "", false, true)
}

func main() {
	/* ------- 初始化 Nacos / 数据库 ------- */
	initNacos()
	initDatabase()
	defer closeDatabase()

	/* ------- Gin 路由 ------- */
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://micro.roliyal.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-User-ID"},
		AllowCredentials: true,
	}))
	router.POST("/login", loginHandler)
	router.POST("/register", registerHandler)
	router.GET("/user", userHandler)

	/* ------- HTTP 服务器 ------- */
	srv := &http.Server{
		Addr:    ":8083",
		Handler: router,
	}

	/* ------- 后台启动 ------- */
	go func() {
		fmt.Println("login-service listening :8083")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	/* ------- 注册实例到 Nacos ------- */
	hostIP, err := getHostIP()
	if err != nil {
		log.Fatalf("get host IP: %v", err)
	}
	if err = registerService(NamingClient, "login-service", hostIP, 8083); err != nil {
		log.Fatalf("register service: %v", err)
	}
	defer deregisterLoginService()

	/* ------- 捕获 SIGTERM / SIGINT ------- */
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Println("received termination signal, shutting down...")

	/* ------- 30 秒内优雅关机 ------- */
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}

	log.Println("server exited gracefully")
}
