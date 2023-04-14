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
	ID        int    `json:"id"` // 使用 'ID' 而不是 'UserID'
}

func main() {
	initNacos() // Initialize Nacos client
	initDatabase()
	defer closeDatabase()

	err := registerService(NamingClient, "login-service", "127.0.0.1", 8083)
	if err != nil {
		fmt.Printf("Error registering game service instance: %v\n", err)
		os.Exit(1)
	}
	defer closeDatabase()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // 允许来自任何域的请求
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
	})

	// 使用 CORS 中间件包装处理程序
	loginHandler := c.Handler(http.HandlerFunc(loginHandler))
	userHandler := c.Handler(http.HandlerFunc(userHandler))
	registerHandler := c.Handler(http.HandlerFunc(registerHandler))

	// 注册处理程序
	http.Handle("/login", loginHandler)
	http.Handle("/user", userHandler)
	http.Handle("/register", registerHandler)

	fmt.Println("Starting server on port 8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
	deregisterGameService()
}
func updateUser(user *User) error {
	if err := db.Model(user).Where("id = ?", user.ID).Update("auth_token", user.AuthToken).Error; err != nil {
		log.Println("Error updating user:", err)
		return err
	}
	return nil
}

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
	db = db.LogMode(true)

	if err := db.Select("ID, Username, Password, AuthToken, Wins, Attempts").Where("username = ? AND password = ?", req.Username, req.Password).First(&user).Error; err == nil {
		log.Println("User found:", user)
		log.Println("User data retrieved from the database:", user)
		log.Println("Generated SQL query:", db.Where("username = ? AND password = ?", req.Username, req.Password).First(&user).QueryExpr())
		fmt.Printf("User data after query: %+v\n", user)

		newAuthToken, err := generateAuthToken()
		if err != nil {
			log.Println("Error generating auth token:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		user.AuthToken = newAuthToken
		fmt.Printf("User data after update: %+v\n", user)

		err = updateUser(&user)
		if err != nil {
			log.Println("Error updating user:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			log.Println("User updated successfully:", user)
		}

		res := loginResponse{
			Success:   true,
			AuthToken: user.AuthToken,
			ID:        user.ID,
		}

		fmt.Println("Updated user:", user)

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

func userHandler(w http.ResponseWriter, r *http.Request) {
	authToken := r.URL.Query().Get("authToken")
	userID := r.URL.Query().Get("userID")

	// 确保userID已提供
	if authToken == "" || userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 在此处添加调试日志
	log.Printf("Received user request with authToken: %s and userID: %s\n", authToken, userID)

	// 使用userID查询用户
	var user User
	if err := db.Where("auth_token = ? AND id = ?", authToken, userID).First(&user).Error; err != nil {
		fmt.Printf("Error finding user by authToken and userID: %v\n", err)
		if gorm.IsRecordNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
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

	var user User
	err = db.Where("username = ?", req.Username).First(&user).Error
	if err == nil {
		log.Println("Username already exists:", req.Username)
		w.WriteHeader(http.StatusConflict)
		return
	}

	if !gorm.IsRecordNotFoundError(err) {
		log.Println("Error checking for existing user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newAuthToken, err := generateAuthToken()
	if err != nil {
		log.Println("Error generating auth token:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user = User{
		Username:  req.Username,
		Password:  req.Password,
		AuthToken: newAuthToken,
		Wins:      0,
		Attempts:  0,
	}

	err = db.Create(&user).Error
	if err != nil {
		log.Println("Error creating new user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
