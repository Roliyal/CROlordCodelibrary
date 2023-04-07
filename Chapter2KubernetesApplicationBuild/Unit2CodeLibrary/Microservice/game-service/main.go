package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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
	initNacos()    // Initialize Nacos client
	initDatabase() // Initialize the database
	defer closeDatabase()

	mux := http.NewServeMux()
	mux.HandleFunc("/game", guessHandler)

	fmt.Println("Starting server on port 8084")
	log.Fatal(http.ListenAndServe(":8084", corsMiddleware(mux)))
}

// Rest of the code...

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
func guessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"number": rand.Intn(100) + 1,
		})
		return
	}

	authToken := extractTokenFromHeader(r)
	userId, err := strconv.Atoi(r.URL.Query().Get("userID"))
	if err != nil {
		log.Println("Error parsing userID:", err)
		w.WriteHeader(http.StatusBadRequest)
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
		} // 保存更新并检查错误       // 保存更新
		//deleteGame(game)
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
	log.Printf("Headers: %v\n", r.Header) // 输出请求头的调试信息

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
