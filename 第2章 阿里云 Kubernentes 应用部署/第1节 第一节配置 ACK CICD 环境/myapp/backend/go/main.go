package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()

	// Serve static files from the "frontend" directory
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend")))

	// Login endpoint
	router.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "admin" && password == "admin" {
			fmt.Fprintln(w, "Login successful")
			return
		}

		http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
	}).Methods("POST")

	// Save score endpoint
	router.HandleFunc("/api/score", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement save score logic
		fmt.Fprintln(w, "Score saved")
	}).Methods("POST")

	// Start server
	log.Println("Server started on port 8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}
