package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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
	// Replace the URL with the actual URL of your login service.
	userID, err := getUserIDFromLoginService(authToken)
	if err != nil {
		return nil, err
	}

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func getUserIDFromLoginService(authToken string) (uint, error) {
	// Implement the actual call to the login service here
	// For now, just return a dummy user ID
	return 1, nil
}

func generateTargetNumber() int {
	// Implement the actual target number generation here
	// For now, just return a dummy target number
	return 42
}
