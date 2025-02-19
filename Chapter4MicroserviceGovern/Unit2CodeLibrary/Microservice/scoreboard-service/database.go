package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
	"time"
)

// User 结构体
type User struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	Password       string    `json:"-"`
	AuthToken      string    `json:"auth_token"`
	Wins           int       `json:"wins"`
	Attempts       int       `json:"attempts"`
	CorrectGuesses int       `json:"correct_guesses"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SetupDatabase 初始化数据库连接
func SetupDatabase(nacosClient config_client.IConfigClient) (*sql.DB, error) {
	dbConfig, err := getDatabaseConfigFromNacos(nacosClient)
	if err != nil {
		return nil, err
	}

	return initDB(dbConfig)
}

// getDatabaseConfigFromNacos 从 Nacos 获取数据库配置
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

// initDB 初始化数据库连接
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

// getScoreboardData 获取排行榜数据
func getScoreboardData(db *sql.DB) ([]ScoreboardEntry, error) {
	query := `
SELECT game.ID, users.username, game.Attempts, game.TargetNumber
    FROM game
    JOIN users ON game.ID = users.id
    ORDER BY game.Attempts ASC
`

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare query: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	var entries []ScoreboardEntry
	for rows.Next() {
		var entry ScoreboardEntry
		err := rows.Scan(&entry.ID, &entry.Username, &entry.Attempts, &entry.TargetNumber)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan row: %v", err)
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Row iteration error: %v", err)
	}

	return entries, nil
}

// closeDatabase 关闭数据库连接
func closeDatabase(db *sql.DB) {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Println("Error closing database:", err)
		}
	}
}
