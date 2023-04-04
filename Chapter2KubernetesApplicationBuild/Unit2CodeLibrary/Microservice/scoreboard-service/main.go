package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
)

type ScoreboardEntry struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
	Attempts int    `json:"attempts"`
}

type getScoreboardResponse struct {
	Entries []ScoreboardEntry `json:"entries"`
}

func main() {
	initNacos()    // Initialize Nacos client
	initDatabase() // Initialize the database
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

	var users []User
	db.Find(&users)

	entries := make([]ScoreboardEntry, len(users))
	for i, user := range users {
		entries[i] = ScoreboardEntry{
			Username: user.Username,
			Wins:     user.Wins,
			Attempts: user.Attempts,
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Wins == entries[j].Wins {
			return entries[i].Attempts < entries[j].Attempts
		}
		return entries[i].Wins > entries[j].Wins
	})

	res := getScoreboardResponse{
		Entries: entries,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
