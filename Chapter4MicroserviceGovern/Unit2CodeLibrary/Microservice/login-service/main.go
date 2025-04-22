package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
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

// ---------- 请求/响应结构 ----------
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

// ---------- 全局变量（假设 initNacos 会设置 NamingClient） ----------
var db *gorm.DB

// ---------- main ----------
func main() {
	// 1. 初始化 Nacos 客户端、数据库
	initNacos()
	initDatabase()
	defer closeDatabase()

	// 2. 获取本机 IP 并注册到 Nacos
	hostIP, err := getHostIP()
	if err != nil {
		log.Fatalf("Failed to get host IP: %v", err)
	}
	if err = registerService(NamingClient, "login-service", hostIP, 8083); err != nil {
		log.Fatalf("register service: %v", err)
	}
	log.Printf(" 已注册 login-service 到 Nacos: %s:8083", hostIP)

	// 3. Gin 路由及 CORS 设置
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://micro.roliyal.com"},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-User-ID"},
	}))

	r.POST("/login", loginHandler)
	r.POST("/register", registerHandler)
	r.GET("/user", userHandler)
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// 4. 用 http.Server 包装 Gin，实现优雅关闭
	srv := &http.Server{
		Addr:    ":8083",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Gin listen error: %v", err)
		}
	}()
	log.Println(" login-service listening on :8083")

	// 5. 信号捕获：收到终止信号后优雅下线
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞直到收到 SIGINT/SIGTERM
	log.Println(" 收到终止信号，开始优雅下线...")

	// 5.1 反注册 Nacos
	if err := deregisterLoginService(); err != nil {
		log.Printf("️ deregisterLoginService error: %v", err)
	} else {
		log.Println("已从 Nacos 注销 login-service")
	}

	// 5.2 优雅关闭 HTTP server，留 20s 给正在处理请求
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf(" HTTP server Shutdown error: %v", err)
	} else {
		log.Println(" HTTP server 已优雅退出")
	}

	log.Println("👋 服务已完全退出")
}

// ---------- token helpers ----------
func generateAuthToken() (string, error) { return generateRandomToken(32) }

func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateToken() string {
	tkn, _ := generateAuthToken()
	return tkn
}

// ---------- handlers ----------
func loginHandler(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"success": false, "error": "invalid method"})
		return
	}

	body, _ := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()

	var req loginRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	var user User
	if err := db.Select("ID, Username, Password, Wins, Attempts, AuthToken").
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
	forceRefresh := c.Query("force") == "true"
	if token == "" || forceRefresh {
		var err error
		token, err = generateAuthToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
			return
		}
		user.AuthToken = token
		if err = db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
	}

	writeAuthCookies(c, token, user.ID)
	c.JSON(http.StatusOK, loginResponse{Success: true, AuthToken: token, ID: user.ID})
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

	if userID == "" && authToken != "" {
		var u User
		if err := db.Select("ID").Where("AuthToken = ?", authToken).First(&u).Error; err == nil {
			userID = u.ID
		}
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

	type userResp struct {
		ID        string    `json:"ID"`
		Username  string    `json:"Username"`
		AuthToken string    `json:"AuthToken"`
		Wins      int       `json:"Wins"`
		Attempts  int       `json:"Attempts"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	c.JSON(http.StatusOK, userResp{
		ID:        user.ID,
		Username:  user.Username,
		AuthToken: user.AuthToken,
		Wins:      user.Wins,
		Attempts:  user.Attempts,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
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

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pwd hash error"})
		return
	}

	user := User{
		ID:        nextID,
		Username:  req.Username,
		Password:  string(hash),
		AuthToken: generateToken(),
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

func writeAuthCookies(c *gin.Context, token, id string) {
	age := 7 * 24 * 3600
	c.SetCookie("AuthToken", token, age, "/", "", false, true)
	c.SetCookie("X-User-ID", id, age, "/", "", false, true)
}
