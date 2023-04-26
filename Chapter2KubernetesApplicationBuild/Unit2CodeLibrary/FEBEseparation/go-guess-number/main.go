package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type GuessRequest struct {
	Guess int `json:"guess"`
}

type GuessResponse struct {
	Message string `json:"message"`
}

type Game struct {
	sync.Mutex
	TargetNumber  int
	LastGuess     int
	LastDirection string
}

func (g *Game) checkGuessHandler(w http.ResponseWriter, r *http.Request) {
	g.Lock()
	defer g.Unlock()

	var req GuessRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var res GuessResponse
	if req.Guess == g.LastGuess {
		res.Message = "您已经尝试过这个数字，请尝试一个不同的数字。"
	} else if req.Guess < g.TargetNumber {
		res.Message = fmt.Sprintf("太小了！试试更大一点的数字（%d ～ 100）。", req.Guess+1)
		g.LastDirection = "up"
	} else if req.Guess > g.TargetNumber {
		res.Message = fmt.Sprintf("太大了！试试更小一点的数字（1 ～ %d）。", req.Guess-1)
		g.LastDirection = "down"
	} else {
		res.Message = "恭喜你，猜对了！"
	}

	g.LastGuess = req.Guess

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())
	targetNumber := rand.Intn(100) + 1
	fmt.Println("Target number:", targetNumber)

	game := &Game{
		TargetNumber: targetNumber,
	}

	router := mux.NewRouter()
	router.Use(corsMiddleware)

	router.HandleFunc("/check-guess", game.checkGuessHandler).Methods("POST", "OPTIONS")

	http.ListenAndServe(":8081", router)
}
