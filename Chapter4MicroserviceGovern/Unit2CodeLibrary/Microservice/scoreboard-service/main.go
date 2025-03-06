package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// ScoreboardEntry 定义了用户在排行榜中的信息
type ScoreboardEntry struct {
	ID           string `json:"id"`
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
	// 加载 .env 文件
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	} else {
		log.Println("Loaded .env file successfully")
	}

	// 输出环境变量的值进行调试
	log.Printf("NACOS_TIMEOUT_MS: %s", os.Getenv("NACOS_TIMEOUT_MS"))
	log.Printf("NACOS_SERVER_PORT: %s", os.Getenv("NACOS_SERVER_PORT"))
	log.Printf("NACOS_SERVER_IP: %s", os.Getenv("NACOS_SERVER_IP"))
}

func main() {
	// 启动 HTTP 服务
	r := gin.Default()

	r.GET("/scoreboard", getScoreboardHandler)

	// 设置优雅关闭
	go func() {
		if err := r.Run(":8085"); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// 监听退出信号并注销服务
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// 注销服务
	err := deregisterService("scoreboard-service", 8085)
	if err != nil {
		log.Fatalf("Error deregistering service: %v", err)
	}
}

// getScoreboardHandler 处理获取排行榜请求
func getScoreboardHandler(c *gin.Context) {
	scoreboardData, err := getScoreboardData(db)
	if err != nil {
		log.Println("Error fetching scoreboard data:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, scoreboardData)
}
