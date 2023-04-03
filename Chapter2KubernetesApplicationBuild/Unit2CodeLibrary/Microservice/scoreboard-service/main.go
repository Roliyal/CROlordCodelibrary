package main

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const serviceName = "scoreboard"
const servicePort = 8085

func main() {
	db, err := SetupDatabase()
	if err != nil {
		log.Fatalf("Failed to set up database: %v", err)
	}

	namingClient, err := createNacosClient()
	if err != nil {
		log.Fatalf("Failed to create Nacos client: %v", err)
	}

	ip := "127.0.0.1"
	port := uint64(servicePort)
	err = registerService(namingClient, serviceName, ip, port)
	if err != nil {
		log.Fatalf("Failed to register service with Nacos: %v", err)
	}

	http.HandleFunc("/scoreboard", func(w http.ResponseWriter, r *http.Request) {
		data, err := getScoreboardData(db)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	srv := &http.Server{
		Addr:    ":8085",
		Handler: nil,
	}

	// 监听退出信号
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		log.Println("Shutting down the server...")

		err = deregisterService(namingClient, serviceName, ip, port)
		if err != nil {
			log.Printf("Failed to deregister service with Nacos: %v", err)
		}

		if err := srv.Shutdown(nil); err != nil {
			log.Printf("Server Shutdown: %v", err)
		}
	}()

	log.Printf("Server is ready to handle requests at %s", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server ListenAndServe: %v", err)
	}
}
