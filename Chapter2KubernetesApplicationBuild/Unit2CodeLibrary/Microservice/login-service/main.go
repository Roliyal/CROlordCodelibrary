package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/rs/cors"
	"io/ioutil"
	"log"
	"net/http"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Success   bool   `json:"success"`
	AuthToken string `json:"authToken"`
	ID        int    `json:"id"` // 使用 'ID' 而不是 'UserID'
}

func main() {
	initNacos()    // Initialize Nacos client
	initDatabase() // Initialize the database
	defer closeDatabase()

	// Register the service with Nacos
	serviceName := "login-service"
	ip := "127.0.0.1"
	port := 8083
	clusterName := "DEFAULT"
	groupName := "DEFAULT_GROUP"

	success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        uint64(port),
		ServiceName: serviceName,
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Metadata:    map[string]string{},
		ClusterName: clusterName,
		GroupName:   groupName,
		Ephemeral:   true,
	})

	if !success || err != nil {
		log.Fatalf("Failed to register service with Nacos: %v", err)
	}

	// Unregister the service when the program exits
	defer func() {
		success, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          ip,
			Port:        uint64(port),
			ServiceName: serviceName,
			Cluster:     clusterName,
			GroupName:   groupName,
			Ephemeral:   true,
		})
		if !success || err != nil {
			log.Printf("Failed to deregister service with Nacos: %v", err)
		}
	}()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // 允许来自任何域的请求
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
	})

	// 使用 CORS 中间件包装处理程序
	loginHandler := c.Handler(http.HandlerFunc(loginHandler))
	userHandler := c.Handler(http.HandlerFunc(userHandler))

	// 注册处理程序
	http.Handle("/login", loginHandler)
	http.Handle("/user", userHandler)

	fmt.Println("Starting server on port 8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
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
