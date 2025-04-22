package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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

// ---------- main ----------
func main() {
	initNacos()
	initDatabase()
	defer closeDatabase()

	hostIP, err := getHostIP()
	if err != nil {
		log.Fatalf("Failed to get host IP: %v", err)
	}
	if err = registerService(NamingClient, "login-service", hostIP, 8083); err != nil {
		log.Fatalf("register service: %v", err)
	}
	defer deregisterLoginService()

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

	fmt.Println("login‑service listening :8083")
	if err := r.Run(":8083"); err != nil {
		log.Fatalf("Gin run: %v", err)
	}
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

	// ------- 密码校验（哈希或旧明文） -------
	passOK := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) == nil ||
		user.Password == req.Password
	if !passOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// ------- 复用或生成 token -------
	token := user.AuthToken

	// 如需“ 强制刷新 token”，前端可加 ?force=true
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

	// ------- 写 cookie & 返回 -------
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

	// ★★★ BEGIN: 用 token 反查 ID（缺 ID 时） ★★★
	if userID == "" && authToken != "" {
		var u User
		if err := db.Select("ID").Where("AuthToken = ?", authToken).First(&u).Error; err == nil {
			userID = u.ID
		}
	}
	// ★★★ END ★★★

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

// ---------- cookie util ----------
func writeAuthCookies(c *gin.Context, token, id string) {
	age := 7 * 24 * 3600
	c.SetCookie("AuthToken", token, age, "/", "", false, true)
	c.SetCookie("X-User-ID", id, age, "/", "", false, true)
}
