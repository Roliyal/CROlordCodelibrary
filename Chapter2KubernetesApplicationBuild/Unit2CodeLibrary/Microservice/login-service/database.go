package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

var db *gorm.DB

func (User) TableName() string {
	return "user"
}

type User struct {
	ID             int       `gorm:"column:ID;primaryKey;autoIncrement"`
	Username       string    `gorm:"column:Username;unique"`
	Password       string    `gorm:"column:Password"`
	AuthToken      string    `gorm:"column:AuthToken"`
	Wins           int       `gorm:"column:Wins"`
	Attempts       int       `gorm:"column:Attempts"`
	AuthTokenExtra string    `gorm:"column:auth_token"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func initDatabase() {
	var err error
	db, err = gorm.Open("mysql", "crolord:RyV3MGZ$@Q5rJ3i^-=@tcp(rm-j6cn3wen02w6f5b94ho.mysql.rds.aliyuncs.com:3306)/crolord?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect to database")
	}

	// 创建数据库表
	db.Table("user").AutoMigrate(&User{})
	// 设置AuthToken的默认值
}

func closeDatabase() {
	db.Close()
}
