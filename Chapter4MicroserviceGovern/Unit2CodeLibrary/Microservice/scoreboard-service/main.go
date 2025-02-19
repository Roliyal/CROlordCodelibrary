package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// ScoreboardEntry 定义了用户在排行榜中的信息
type ScoreboardEntry struct {
	ID           string `json:"id"` // 改为 string 类型
	Username     string `json:"username"`
	Attempts     int    `json:"attempts"`
	TargetNumber int    `json:"target_number"`
}

// getScoreboardResponse 定义了返回的 JSON 数据结构
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

// main 启动 HTTP 服务
func main() {
	_, configClient, err := initNacos()
	if err != nil {
		log.Fatal("Error initializing Nacos:", err)
	}
	defer func() {
		err = deregisterService("scoreboard-service", 8085)
		if err != nil {
			log.Fatal("Error deregistering service:", err)
		}
	}()

	db, err = SetupDatabase(configClient)
	if err != nil {
		log.Fatal("Error setting up the database:", err)
	}
	defer closeDatabase(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/scoreboard", getScoreboardHandler)

	fmt.Println("Starting server on port 8085")
	log.Fatal(http.ListenAndServe(":8085", corsMiddleware(mux)))
}

// corsMiddleware 设置 CORS 头
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getScoreboardHandler 处理获取排行榜的 HTTP 请求
func getScoreboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scoreboardData)
}
