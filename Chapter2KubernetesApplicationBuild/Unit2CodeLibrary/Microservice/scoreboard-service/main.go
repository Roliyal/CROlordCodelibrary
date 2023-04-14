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

type ScoreboardEntry struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
	Attempts int    `json:"attempts"`
}

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

type getScoreboardResponse struct {
	Entries []ScoreboardEntry `json:"entries"`
}

var db *sql.DB

func main() {
	nacosClient, configClient, err := initNacos()
	if err != nil {
		log.Fatal("Error initializing Nacos:", err)
	}
	defer func() {
		err = deregisterService(nacosClient, "scoreboard-service", "localhost", 8085)
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

func getScoreboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	gameData, err := getGameData(db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gameData)
}

func closeDatabase() {
	if db != nil {
		db.Close()
	}
}
