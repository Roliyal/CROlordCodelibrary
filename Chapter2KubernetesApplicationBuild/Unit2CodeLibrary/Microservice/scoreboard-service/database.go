package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Score struct {
	Username string `json:"username"`
	Score    int    `json:"score"`
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

func getScoreboardData(db *sql.DB) ([]Score, error) {
	rows, err := db.Query("SELECT username, score FROM scores ORDER BY score DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []Score
	for rows.Next() {
		var score Score
		err = rows.Scan(&score.Username, &score.Score)
		if err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return scores, nil
}
