package main

import (
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type guessRequest struct {
	AuthToken string `json:"authToken"`
	Number    int    `json:"number"`
}

type guessResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Attempts int    `json:"attempts"`
}

func main() {
	logDir := "/app/log"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, 0777)
	}
	initNacos()
	err := registerService(NamingClient, "game-service", "127.0.0.1", 8084)
	if err != nil {
		fmt.Printf("Error registering game service instance: %v\n", err)
		os.Exit(1)
	}
	subscribeLoginService()

	dbConfig, err := getDatabaseConfigFromNacos()
	if err != nil {
		panic("failed to get database configuration from Nacos")
	}
	initDatabase(dbConfig) // Initialize the database with the configuration from Nacos
	defer closeDatabase()

	mux := http.NewServeMux()
	mux.HandleFunc("/game", guessHandler)
	mux.HandleFunc("/health", healthCheckHandler)

	fmt.Println("Starting server on port 8084")
	log.Fatal(http.ListenAndServe(":8084", corsMiddleware(mux)))

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// 注销 game 服务实例
	deregisterGameService()
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func getDatabaseConfigFromNacos() (map[string]string, error) {
	DataId := "Prod_DATABASE"
	Group := "DEFAULT_GROUP"

	fmt.Printf("Requesting Nacos config with DataId: %s, Group: %s\n", DataId, Group) // 输出请求的 DataId 和 Group

	config, err := ConfigClient.GetConfig(vo.ConfigParam{
		DataId: DataId,
		Group:  Group,
	})
	if err != nil {
		return nil, err
	}

	fmt.Printf("Received Nacos config: %s\n", config) // 输出从 Nacos 接收到的配置

	var dbConfig map[string]string
	err = json.Unmarshal([]byte(config), &dbConfig)
	if err != nil {
		return nil, err
	}

	return dbConfig, nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置允许的来源
		w.Header().Set("Access-Control-Allow-Origin", "http://micro.roliyal.com")

		// 设置允许的请求头，包括自定义头 'X-User-ID'
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")

		// 设置允许的HTTP方法
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// 允许携带凭证
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// 处理预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 调用下一个处理器
		next.ServeHTTP(w, r)
	})
}
func guessHandler(w http.ResponseWriter, r *http.Request) {
	// 输出请求头，确保 Authorization 和 X-User-ID 被接收到
	log.Printf("Received headers: %v", r.Header)

	authToken := extractTokenFromHeader(r)
	userIdStr := r.Header.Get("X-User-ID") // 读取 X-User-ID 请求头

	if userIdStr == "" {
		log.Println("Error: Missing X-User-ID header")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Missing X-User-ID header",
		})
		return
	}

	// 转换 userIdStr 为整数
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		log.Println("Error parsing userID:", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid userID",
		})
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req guessRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.AuthToken = authToken
	user, err := getUserFromAuthToken(req.AuthToken, uint(userId)) // 使用 userId 变量
	if err != nil {
		log.Printf("Error getting user from auth token: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	game, err := getOrCreateGame(&user)
	if err != nil {
		log.Println("Error getting or creating game:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var res guessResponse
	if req.Number == game.TargetNumber {
		res.Success = true
		res.Message = "Congratulations! You guessed the correct number."
		res.Attempts = game.Attempts
		game.CorrectGuesses++ // 增加猜中次数
		if err := db.Save(game).Error; err != nil {
			log.Printf("Error updating game: %v", err)
		}
	} else {
		res.Success = false
		if req.Number < game.TargetNumber {
			res.Message = "The number is too low."
		} else {
			res.Message = "The number is too high."
		}
		incrementAttempts(game)
		res.Attempts = game.Attempts
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func extractTokenFromHeader(r *http.Request) string {
	log.Printf("Headers: %v\n", r.Header) // 输出请求头的调试

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 {
		return ""
	}
	return bearerToken[1]
}
