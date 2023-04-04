package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/satori/go.uuid"
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
	AuthToken string `json:"authToken,omitempty"`
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

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	//http.HandleFunc("/user", getUserHandler)
	http.HandleFunc("/user", userHandler) // 添加这行代码

	fmt.Println("Starting server on port 8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}

func updateUser(user *User) error {
	if err := db.Model(user).UpdateColumn("auth_token", user.AuthToken).Error; err != nil {
		log.Println("Error updating user:", err)
		return err
	}
	return nil
}

func generateAuthToken() string {
	u := uuid.NewV4()
	return u.String()
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

	var user User
	if err := db.Where("username = ? AND password = ?", req.Username, req.Password).First(&user).Error; err == nil {
		log.Println("User found:", user)

		if user.AuthToken == "" {
			user.AuthToken = generateAuthToken()
		}
		err := updateUser(&user)

		if err != nil {
			log.Println("Error updating user:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			log.Println("User updated successfully:", user)
		}

		fmt.Println("Updated user:", user)

		res := loginResponse{
			Success:   true,
			AuthToken: user.AuthToken,
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

	var req loginRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u := uuid.NewV4()
	authToken := u.String()

	user := User{
		Username:  req.Username,
		Password:  req.Password,
		AuthToken: authToken,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Println("Error creating user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("User created:", user)

	res := loginResponse{
		Success:   true,
		AuthToken: authToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	fmt.Println("Sent register response:", res)
}
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	authToken := r.URL.Query().Get("authToken")

	var user User
	if err := db.Where("auth_token = ?", authToken).First(&user).Error; err != nil {
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
func userHandler(w http.ResponseWriter, r *http.Request) {
	authToken := r.URL.Query().Get("authToken")
	if authToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user User
	if err := db.Where("auth_token = ?", authToken).First(&user).Error; err != nil {
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
