// main.go

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/rs/cors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time" // 添加 time 包

	// 如果使用密码哈希，请取消注释以下导入
	// "golang.org/x/crypto/bcrypt"
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
	ID        uint   `json:"id"` // 使用 uint 类型
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

	// 配置 CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://micro.roliyal.com"}, // 明确指定前端地址
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},    // 包含 OPTIONS 方法
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-User-ID"}, // 明确列出允许的请求头
		Debug:            true,                                                   // 启用调试日志
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/user", userHandler)
	mux.HandleFunc("/register", registerHandler)

	// 应用 CORS 中间件到整个 ServeMux
	handler := c.Handler(mux)

	fmt.Println("Starting server on port 8083")
	log.Fatal(http.ListenAndServe(":8083", handler))
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
func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received login request")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req loginRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Received login request with username: %s, password: %s\n", req.Username, req.Password)

	var user User
	// db 已在 database.go 中定义为全局变量
	if err := db.Select("ID, Username, Password, Wins, Attempts, AuthToken").Where("username = ?", req.Username).First(&user).Error; err == nil {
		log.Println("User found:", user)

		// 如果使用密码哈希，请在这里验证密码
		/*
		   if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		       // 密码不匹配
		       res := loginResponse{
		           Success: false,
		       }
		       w.Header().Set("Content-Type", "application/json")
		       w.WriteHeader(http.StatusOK)
		       json.NewEncoder(w).Encode(res)
		       fmt.Println("Sent login response:", res)
		       return
		   }
		*/

		// 生成新的 AuthToken
		newAuthToken, err := generateAuthToken()
		if err != nil {
			log.Println("Error generating auth token:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 更新用户的 AuthToken
		user.AuthToken = newAuthToken
		if err := db.Save(&user).Error; err != nil {
			log.Println("Error updating user with new AuthToken:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := loginResponse{
			Success:   true,
			AuthToken: newAuthToken,
			ID:        user.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		fmt.Println("Sent login response:", res)
	} else {
		log.Println("User not found, error:", err)
		res := loginResponse{
			Success: false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		fmt.Println("Sent login response:", res)
	}
}

// userHandler 处理获取用户信息的请求
func userHandler(w http.ResponseWriter, r *http.Request) {
	authToken := r.Header.Get("Authorization")
	userID := r.Header.Get("X-User-ID") // 获取 X-User-ID 请求头

	if authToken == "" || userID == "" {
		log.Println("Error: Missing Authorization or X-User-ID header")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Missing Authorization or X-User-ID header",
		})
		return
	}

	// 将 userID 转换为 uint64
	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		log.Println("Error parsing userID:", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid userID",
		})
		return
	}

	// 使用 authToken 和 userID 查询用户
	var user User
	if err := db.Where("AuthToken = ? AND ID = ?", authToken, userIDUint).First(&user).Error; err != nil {
		log.Printf("Error finding user by AuthToken and ID: %v\n", err)
		if gorm.IsRecordNotFoundError(err) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Unauthorized",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Internal Server Error",
			})
		}
		return
	}

	// 返回用户数据（不包括密码）
	type userResponse struct {
		ID        uint      `json:"ID"`
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// registerHandler 处理用户注册请求
func registerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received register request")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req registerRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Received register request with username: %s, password: %s\n", req.Username, req.Password)

	var existingUser User
	err = db.Where("Username = ?", req.Username).First(&existingUser).Error
	if err == nil {
		log.Println("Username already exists:", req.Username)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Username already exists",
		})
		return
	}

	if !gorm.IsRecordNotFoundError(err) {
		log.Println("Error checking for existing user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Internal Server Error",
		})
		return
	}

	// 如果使用密码哈希，请在这里哈希密码
	/*
	   hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	   if err != nil {
	       log.Println("Error hashing password:", err)
	       w.WriteHeader(http.StatusInternalServerError)
	       json.NewEncoder(w).Encode(map[string]interface{}{
	           "success": false,
	           "error":   "Internal Server Error",
	       })
	       return
	   }
	*/

	user := User{
		Username:  req.Username,
		Password:  req.Password, // 如果使用哈希密码，请设置为 string(hashedPassword)
		AuthToken: generateToken(),
		Wins:      0,
		Attempts:  0,
	}

	err = db.Create(&user).Error
	if err != nil {
		log.Println("Error creating new user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Internal Server Error",
		})
		return
	}

	res := loginResponse{
		Success:   true,
		AuthToken: user.AuthToken,
		ID:        user.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
	fmt.Println("Sent register response:", res)
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
