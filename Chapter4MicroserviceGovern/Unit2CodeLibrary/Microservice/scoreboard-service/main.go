// main.go
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

// SetupDatabase 设置数据库连接
func SetupDatabase(nacosClient config_client.IConfigClient) (*sql.DB, error) {
	dbConfig, err := getDatabaseConfigFromNacos(nacosClient)
	if err != nil {
		return nil, err
	}

	return initDB(dbConfig)
}

// getDatabaseConfigFromNacos 从 Nacos 配置中心获取数据库配置
func getDatabaseConfigFromNacos(nacosClient config_client.IConfigClient) (map[string]string, error) {
	content, err := nacosClient.GetConfig(vo.ConfigParam{
		DataId: "Prod_DATABASE",
		Group:  "DEFAULT_GROUP",
	})

	if err != nil {
		return nil, err
	}

	var dbConfig map[string]string
	err = json.Unmarshal([]byte(content), &dbConfig)
	if err != nil {
		return nil, err
	}

	return dbConfig, nil
}

// initDB 初始化数据库连接
func initDB(dbConfig map[string]string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True",
		dbConfig["DB_USER"], dbConfig["DB_PASSWORD"], dbConfig["DB_HOST"], dbConfig["DB_PORT"], dbConfig["DB_NAME"])

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// getScoreboardData 从数据库获取排行榜数据
func getScoreboardData(db *sql.DB) ([]ScoreboardEntry, error) {
	query := `
        SELECT game.id, users.username, game.attempts, game.target_number
        FROM game
        JOIN users ON game.user_id = users.id
        ORDER BY game.attempts DESC
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ScoreboardEntry
	for rows.Next() {
		var entry ScoreboardEntry
		err := rows.Scan(&entry.ID, &entry.Username, &entry.Attempts, &entry.TargetNumber)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// closeDatabase 关闭数据库连接
func closeDatabase() {
	if db != nil {
		db.Close()
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
	defer closeDatabase()

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
