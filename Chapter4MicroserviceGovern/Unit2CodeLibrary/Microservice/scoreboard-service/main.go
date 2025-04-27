// main.go
package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

// ScoreboardEntry 定义了用户在排行榜中的信息
type ScoreboardEntry struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Attempts     int    `json:"attempts"`
	TargetNumber int    `json:"target_number"`
}

var db *sql.DB
var zapLog *zap.SugaredLogger

// ---------- Logger ----------
func initLogger() {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	logger, _ := cfg.Build()
	zapLog = logger.Sugar()
}

// ---------- .env ----------
func init() {
	initLogger()
	pwd, err := os.Getwd()
	if err != nil {
		zapLog.Fatalf("Could not determine working directory: %v", err)
	}
	if err = godotenv.Load(filepath.Join(pwd, ".env")); err != nil {
		zapLog.Fatalf("Error loading .env file: %v", err)
	}
}

// ---------- main ----------
func main() {
	_, configClient, err := initNacos()
	if err != nil {
		zapLog.Fatal("Error initializing Nacos:", err)
	}
	defer func() {
		if err = deregisterService("scoreboard-service", 8085); err != nil {
			zapLog.Fatal("Error deregistering service:", err)
		}
	}()

	db, err = SetupDatabase(configClient)
	if err != nil {
		zapLog.Fatal("Error setting up the database:", err)
	}
	defer closeDatabase(db)

	r := gin.Default()
	r.Use(corsMiddleware)
	r.GET("/scoreboard", getScoreboardHandler)

	zapLog.Infof("Starting server on port 8085")
	if err = r.Run(":8085"); err != nil {
		zapLog.Fatal("Error starting server:", err)
	}
}

// ---------- CORS ----------
func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "http://micro.roliyal.com")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Max-Age", "100")
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}
	c.Next()
}

// ---------- Handler ----------
func getScoreboardHandler(c *gin.Context) {
	data, err := getScoreboardData(db)
	if err != nil {
		zapLog.Errorw("Error fetching scoreboard data", "err", err)
		c.JSON(500, gin.H{"error": "Internal Server Error", "success": false})
		return
	}
	c.JSON(200, data)
}
