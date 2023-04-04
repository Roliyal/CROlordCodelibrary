package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io/ioutil"
	"net/http"
)

var db *gorm.DB

type User struct {
	ID       uint   `gorm:"primary_key"`
	Username string `gorm:"unique"`
	Password string
}

type Game struct {
	ID           uint `gorm:"primary_key"`
	UserID       uint
	TargetNumber int
	Attempts     int
}

func initDatabase() {
	var err error
	db, err = gorm.Open("mysql", "crolord:RyV3MGZ$@Q5rJ3i^-=@tcp(rm-j6cn3wen02w6f5b94ho.mysql.rds.aliyuncs.com:3306)/crolord?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect to database")
	}

	// 创建数据库表
	db.AutoMigrate(&User{}, &Game{})
}

func closeDatabase() {
	db.Close()
}

func getOrCreateGame(user *User) (*Game, error) {
	var game Game
	if err := db.Where("user_id = ?", user.ID).First(&game).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			game.UserID = user.ID
			game.TargetNumber = generateTargetNumber()
			game.Attempts = 0
			if err := db.Create(&game).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &game, nil
}

func incrementAttempts(game *Game) {
	game.Attempts++
	db.Save(game)
}

func deleteGame(game *Game) {
	db.Delete(game)
}

func getUserFromAuthToken(authToken string) (*User, error) {
	// Call the login service to get the user ID from the auth token.
	userID, err := getUserIDFromLoginService(authToken)
	if err != nil {
		return nil, err
	}
	fmt.Printf("User ID from login service: %d\n", userID) // 添加这行代码

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func getUserIDFromLoginService(authToken string) (uint, error) {
	loginServiceURL := getLoginServiceURL()
	requestURL := fmt.Sprintf("%s/user?authToken=%s", loginServiceURL, authToken)
	fmt.Printf("Requesting user ID with URL: %s\n", requestURL) //

	resp, err := http.Get(requestURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Response body from login service: %s\n", string(respBody))
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("login service returned status %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	return uint(data["id"].(float64)), nil
}

func generateTargetNumber() int {
	// Implement the actual target number generation here
	// For now, just return a dummy target number
	return 42
}
