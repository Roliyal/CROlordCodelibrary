package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"log"
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

	r := gin.Default()

	// 处理 CORS 请求
	r.Use(corsMiddleware)

	// 路由处理
	r.GET("/scoreboard", getScoreboardHandler)

	fmt.Println("Starting server on port 8085")
	if err := r.Run(":8085"); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

// corsMiddleware 设置 CORS 头
func corsMiddleware(c *gin.Context) {
	// 设置允许的源（特定的域名，不能是 *）
	c.Header("Access-Control-Allow-Origin", "http://micro.roliyal.com")

	// 允许的请求头
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")

	// 允许的请求方法
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	// 允许携带凭证（如 cookies）
	c.Header("Access-Control-Allow-Credentials", "true")

	// 设置预检请求的缓存时间
	c.Header("Access-Control-Max-Age", "100")

	// 如果是预检请求，直接返回 200
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}

	// 继续处理请求
	c.Next()
}

// getScoreboardHandler 处理获取排行榜的 HTTP 请求
func getScoreboardHandler(c *gin.Context) {
	scoreboardData, err := getScoreboardData(db)
	if err != nil {
		log.Println("Error fetching scoreboard data:", err)
		c.JSON(500, gin.H{
			"error":   "Internal Server Error",
			"success": false,
		})
		return
	}

	c.JSON(200, scoreboardData)
}
