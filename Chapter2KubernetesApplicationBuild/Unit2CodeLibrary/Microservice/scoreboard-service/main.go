package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
	"net/http"
)

type ScoreboardEntry struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
	Attempts int    `json:"attempts"`
}

type getScoreboardResponse struct {
	Entries []ScoreboardEntry `json:"entries"`
}

var db *sql.DB

func main() {
	initNacos()
	db = initDatabase()
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

func initNacos() {
	nacosClient, err := createNacosClient()
	if err != nil {
		log.Fatal("Error creating Nacos client:", err)
	}

	err = registerService(nacosClient, "scoreboard-service", "localhost", 8085)
	if err != nil {
		log.Fatal("Error registering service:", err)
	}

	defer func() {
		err = deregisterService(nacosClient, "scoreboard-service", "localhost", 8085)
		if err != nil {
			log.Fatal("Error deregistering service:", err)
		}
	}()
}

func initDatabase() *sql.DB {
	db, err := setupDatabase()
	if err != nil {
		log.Fatal("Error setting up the database:", err)
	}
	return db
}

func closeDatabase() {
	if db != nil {
		db.Close()
	}
}

func setupDatabase() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True", dbUser, dbPassword, dbHost, dbName)
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

func createNacosClient() (naming_client.INamingClient, error) {
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: "mse-c00253114-p.nacos-ans.mse.aliyuncs.com",
			Port:   80,
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           5000,
		LogDir:              "nacos-log",
		CacheDir:            "nacos-cache",
		UpdateThreadNum:     2,
		NotLoadCacheAtStart: true,
	}

	nacosClient, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})

	return nacosClient, err
}

func registerService(client naming_client.INamingClient, serviceName, ip string, port uint64) error {
	success, err := client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	})

	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("Failed to register service")
	}

	return nil
}

func deregisterService(client naming_client.INamingClient, serviceName, ip string, port uint64) error {
	success, err := client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		Ephemeral:   true,
	})

	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("Failed to deregister service")
	}

	return nil
}
