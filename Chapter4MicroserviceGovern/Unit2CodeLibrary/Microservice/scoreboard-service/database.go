package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"time"
)

// database.go

type User struct {
	ID             uint   `gorm:"primary_key"`
	Username       string `gorm:"unique"`
	Password       string
	AuthToken      string
	Wins           int
	Attempts       int
	CorrectGuesses int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ScoreboardEntry struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Attempts     int    `json:"attempts"`
	TargetNumber int    `json:"target_number"`
}

func SetupDatabase(nacosClient config_client.IConfigClient) (*sql.DB, error) {
	dbConfig, err := getDatabaseConfigFromNacos(nacosClient)
	if err != nil {
		return nil, err
	}

	return initDB(dbConfig)
}

func getDatabaseConfigFromNacos(nacosClient config_client.IConfigClient) (map[string]string, error) {
	content, err := nacosClient.GetConfig(vo.ConfigParam{
		DataId: "Prod_DATABASE",
		Group:  "DEFAULT_GROUP",
	})

	if err != nil {
		return nil, err
	}

	var dbConfig map[string]string
	err = json.Unmarshal([]byte(content), &dbConfig)
	if err != nil {
		return nil, err
	}

	return dbConfig, nil
}

func initDB(dbConfig map[string]string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True",
		dbConfig["DB_USER"], dbConfig["DB_PASSWORD"], dbConfig["DB_HOST"], dbConfig["DB_PORT"], dbConfig["DB_NAME"])

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

func getScoreboardData(db *sql.DB) ([]ScoreboardEntry, error) {
	query := `
        SELECT game.id, users.username, game.attempts, game.target_number
        FROM game
        JOIN users ON game.id = users.id
        ORDER BY game.attempts DESC
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ScoreboardEntry
	for rows.Next() {
		var entry ScoreboardEntry
		err := rows.Scan(&entry.ID, &entry.Username, &entry.Attempts, &entry.TargetNumber)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
