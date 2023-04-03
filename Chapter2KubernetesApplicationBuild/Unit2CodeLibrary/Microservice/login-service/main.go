package main

import (
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"io/ioutil"
	"log"
	"net/http"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Success bool `json:"success"`
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

	fmt.Println("Starting server on port 8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
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

	res := loginResponse{}

	var user User
	if err := db.Where("username = ? AND password = ?", req.Username, req.Password).First(&user).Error; err == nil {
		log.Println("User found:", user)
		res.Success = true
	} else {
		log.Println("User not found, error:", err)
		res.Success = false
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Implement registration logic here
}
