package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type GameData struct {
	ID           int `json:"id"`
	Attempts     int `json:"attempts"`
	TargetNumber int `json:"target_number"`
}

const (
	dbHost     = "rm-j6cn3wen02w6f5b94ho.mysql.rds.aliyuncs.com"
	dbUser     = "crolord"
	dbPassword = "RyV3MGZ$@Q5rJ3i^-="
	dbName     = "crolord"
)

func SetupDatabase() (*sql.DB, error) {
	return initDB()
}

func initDB() (*sql.DB, error) {
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
