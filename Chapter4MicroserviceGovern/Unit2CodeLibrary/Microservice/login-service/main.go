package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

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

type userResponse struct {
	ID        string    `json:"ID"`
	Username  string    `json:"Username"`
	AuthToken string    `json:"AuthToken"`
	Wins      int       `json:"Wins"`
	Attempts  int       `json:"Attempts"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func main() {
	initNacos()
	initDatabase()
	defer closeDatabase()

	hostIP, err := getHostIP()
	if err != nil {
		log.Fatalf("Failed to get host IP: %v", err)
	}

	err = registerService(NamingClient, "login-service", hostIP, 8083)
	if err != nil {
		fmt.Printf("Error registering login service instance: %v\n", err)
		os.Exit(1)
	}
	defer deregisterLoginService()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://micro.roliyal.com"},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-User-ID"},
	}))

	r.POST("/login", loginHandler)
	r.POST("/register", registerHandler)
	r.GET("/user", userHandler)

	fmt.Println("Starting server on port 8083")
	r.Run(":8083")
}

func loginHandler(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"success": false, "error": "Invalid method"})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}
	defer c.Request.Body.Close()

	var req loginRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	var user User
	if err := db.Select("ID, Username, Password, Wins, Attempts, AuthToken").Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, loginResponse{Success: false})
		return
	}

	if user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid password"})
		return
	}

	newAuthToken, err := generateAuthToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating auth token"})
		return
	}

	user.AuthToken = newAuthToken
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating auth token"})
		return
	}

	// ✅ 设置 Cookie
	c.SetCookie("X-User-ID", user.ID, 3600, "/", "micro.roliyal.com", false, true)

	c.JSON(http.StatusOK, loginResponse{
		Success:   true,
		AuthToken: newAuthToken,
		ID:        user.ID,
	})
}

func registerHandler(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	var existingUser User
	err := db.Where("Username = ?", req.Username).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	if !gorm.IsRecordNotFoundError(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	nextID, err := getNextUserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID generation failed"})
		return
	}

	user := User{
		ID:        nextID,
		Username:  req.Username,
		Password:  req.Password,
		AuthToken: generateToken(),
		Wins:      0,
		Attempts:  0,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating new user"})
		return
	}

	c.JSON(http.StatusCreated, loginResponse{
		Success:   true,
		AuthToken: user.AuthToken,
		ID:        user.ID,
	})
}

func userHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		cookieUserID, _ := c.Cookie("X-User-ID")
		userID = cookieUserID
	}

	if authToken == "" || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization or X-User-ID"})
		return
	}

	var user User
	if err := db.Where("AuthToken = ? AND ID = ?", authToken, userID).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
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

// 工具函数

func generateAuthToken() (string, error) {
	return generateRandomToken(32)
}

func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateToken() string {
	token, err := generateRandomToken(32)
	if err != nil {
		log.Println("Error generating token:", err)
		return ""
	}
	return token
}
