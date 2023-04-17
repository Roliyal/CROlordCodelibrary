package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type GameData struct {
	ID           int `json:"id"`
	Attempts     int `json:"attempts"`
	TargetNumber int `json:"target_number"`
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
func getGameData(db *sql.DB) ([]GameData, error) {
	rows, err := db.Query("SELECT id, attempts, target_number FROM game")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gameData []GameData
	for rows.Next() {
		var data GameData
		err = rows.Scan(&data.ID, &data.Attempts, &data.TargetNumber)
		if err != nil {
			return nil, err
		}
		gameData = append(gameData, data)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return gameData, nil
}
