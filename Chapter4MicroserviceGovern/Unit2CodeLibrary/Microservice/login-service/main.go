package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors" // 使用 Gin 专用的 CORS 中间件
	"github.com/gin-gonic/gin"    // 引入 Gin 框架
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time" // 添加 time 包
)

// 定义请求结构体
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
	AuthToken string `json:"authToken"`
	ID        string `json:"id"`
}

func main() {
	initNacos()    // 初始化 Nacos 客户端
	initDatabase() // 初始化数据库连接
	defer closeDatabase()

	// 获取主机 IP 地址
	hostIP, err := getHostIP()
	if err != nil {
		log.Fatalf("Failed to get host IP: %v", err)
	}

	// 注册服务到 Nacos
	err = registerService(NamingClient, "login-service", hostIP, 8083)
	if err != nil {
		fmt.Printf("Error registering login service instance: %v\n", err)
		os.Exit(1)
	}
	defer deregisterLoginService() // 确保 deregisterLoginService 已定义

	// 使用 Gin 创建一个 HTTP 引擎
	r := gin.Default()

	// 配置 CORS（使用 Gin 专用的 CORS 中间件）
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://micro.roliyal.com"}, // 明确指定前端地址
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},    // 包含 OPTIONS 方法
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-User-ID"}, // 明确列出允许的请求头
		AllowOriginFunc: func(origin string) bool {
			if origin == "" {
				// 允许无 Origin 的请求（服务器间请求）
				return true
			}
			for _, o := range []string{"http://micro.roliyal.com"} {
				if o == origin {
					return true
				}
			}
			return false
		},
		Debug: true, // 启用调试日志
	}))

	// 定义路由
	r.POST("/login", loginHandler)
	r.POST("/register", registerHandler)
	r.GET("/user", userHandler)

	// 启动服务
	fmt.Println("Starting server on port 8083")
	r.Run(":8083")
}

// generateAuthToken 生成认证令牌
func generateAuthToken() (string, error) {
	return generateRandomToken(32)
}

// generateRandomToken 生成指定长度的随机令牌
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// loginHandler 处理登录请求
func loginHandler(c *gin.Context) {
	fmt.Println("Received login request")
	if c.Request.Method != "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"success": false, "error": "Invalid method"})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}
	defer c.Request.Body.Close()

	var req loginRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	log.Printf("Received login request with username: %s\n", req.Username)

	var user User
	// db 已在 database.go 中定义为全局变量
	if err := db.Select("ID, Username, Password, Wins, Attempts, AuthToken").Where("username = ?", req.Username).First(&user).Error; err == nil {
		log.Println("User found:", user)

		// 生成新的 AuthToken
		newAuthToken, err := generateAuthToken()
		if err != nil {
			log.Println("Error generating auth token:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating auth token"})
			return
		}

		// 更新用户的 AuthToken
		user.AuthToken = newAuthToken
		if err := db.Save(&user).Error; err != nil {
			log.Println("Error updating user with new AuthToken:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating auth token"})
			return
		}

		res := loginResponse{
			Success:   true,
			AuthToken: newAuthToken,
			ID:        user.ID, // 现在是string类型
		}

		c.JSON(http.StatusOK, res)
	} else {
		log.Println("User not found, error:", err)
		res := loginResponse{
			Success: false,
		}

		c.JSON(http.StatusOK, res)
	}
}

// userHandler 处理获取用户信息的请求
func userHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	userID := c.GetHeader("X-User-ID") // 获取 X-User-ID 请求头

	log.Printf("Received headers: Authorization=%s, X-User-ID=%s", authToken, userID)

	if authToken == "" || userID == "" {
		log.Println("Error: Missing Authorization or X-User-ID header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization or X-User-ID header"})
		return
	}

	// 使用 authToken 和 userID 查询用户
	var user User
	if err := db.Where("AuthToken = ? AND ID = ?", authToken, userID).First(&user).Error; err != nil {
		log.Printf("Error finding user by AuthToken and ID: %v\n", err)
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		}
		return
	}

	// 返回用户数据（不包括密码）
	type userResponse struct {
		ID        string    `json:"ID"`
		Username  string    `json:"Username"`
		AuthToken string    `json:"AuthToken"`
		Wins      int       `json:"Wins"`
		Attempts  int       `json:"Attempts"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	res := userResponse{
		ID:        user.ID,
		Username:  user.Username,
		AuthToken: user.AuthToken,
		Wins:      user.Wins,
		Attempts:  user.Attempts,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, res)
}

// registerHandler 处理用户注册请求
func registerHandler(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	log.Printf("Received register request with username: %s\n", req.Username)

	var existingUser User
	err := db.Where("Username = ?", req.Username).First(&existingUser).Error
	if err == nil {
		log.Println("Username already exists:", req.Username)
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	if !gorm.IsRecordNotFoundError(err) {
		log.Println("Error checking for existing user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	nextID, err := getNextUserID()
	if err != nil {
		log.Println("Error generating User ID:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	user := User{
		ID:        nextID,
		Username:  req.Username,
		Password:  req.Password, // 可替换为密码哈希
		AuthToken: generateToken(),
		Wins:      0,
		Attempts:  0,
	}

	err = db.Create(&user).Error
	if err != nil {
		log.Println("Error creating new user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	res := loginResponse{
		Success:   true,
		AuthToken: user.AuthToken,
		ID:        user.ID,
	}

	c.JSON(http.StatusCreated, res)
}

// generateToken 生成认证令牌（简单示例）
func generateToken() string {
	token, err := generateRandomToken(32)
	if err != nil {
		log.Println("Error generating token:", err)
		return ""
	}
	return token
}
