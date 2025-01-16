// main.go

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type getScoreboardResponse struct {
	Entries []ScoreboardEntry `json:"entries"`
}

var db *sql.DB

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could not determine working directory: %v", err)
	}
	envPath := filepath.Join(pwd, ".env")
	err = godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading .env file from %s: %v", envPath, err)
	}
}

func main() {
	// 初始化 Nacos 客户端并获取配置客户端
	namingClient, configClient, err := initNacos()
	if err != nil {
		log.Fatalf("Error initializing Nacos: %v", err)
	}
	defer func() {
		err = deregisterService("scoreboard-service", 8085)
		if err != nil {
			log.Fatalf("Error deregistering service: %v", err)
		}
	}()

	// 设置数据库连接
	db, err = SetupDatabase(configClient)
	if err != nil {
		log.Fatalf("Error setting up the database: %v", err)
	}
	defer closeDatabase()

	// 设置 HTTP 路由
	mux := http.NewServeMux()
	mux.HandleFunc("/scoreboard", getScoreboardHandler)

	// 应用 CORS 中间件
	handler := corsMiddleware(mux)

	// 启动 HTTP 服务器
	fmt.Println("Starting server on port 8085")
	log.Fatal(http.ListenAndServe(":8085", handler))
}

// corsMiddleware CORS 中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置允许的来源
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// 设置允许的请求头，包括自定义头 'Authorization'
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

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

// getScoreboardHandler 处理排行榜请求
func getScoreboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Method not allowed",
			"success": false,
		})
		return
	}

	scoreboardData, err := getScoreboardData(db)
	if err != nil {
		log.Println("Error fetching scoreboard data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Internal Server Error",
			"success": false,
		})
		return
	}

	response := getScoreboardResponse{
		Entries: scoreboardData,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Error encoding JSON response:", err)
	}
}
