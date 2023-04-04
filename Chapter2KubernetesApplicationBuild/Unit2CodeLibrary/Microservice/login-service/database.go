package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

func (User) TableName() string {
	return "user"
}

type User struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Username  string `json:"username" gorm:"size:255;unique;not null"`
	Password  string `json:"-" gorm:"size:255;not null"`
	AuthToken string `json:"auth_token,omitempty" gorm:"size:255;default:''"`
	Wins      uint   `json:"wins"`
	Attempts  uint   `json:"attempts"`
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
	db.Exec("ALTER TABLE `user` ALTER `AuthToken` SET DEFAULT ''")
}

func closeDatabase() {
	db.Close()
}
