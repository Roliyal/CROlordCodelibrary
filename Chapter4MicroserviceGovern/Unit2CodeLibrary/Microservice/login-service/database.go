package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"go.uber.org/zap"
)

/* ----------------- 数据模型 ----------------- */

type User struct {
	ID             string    `gorm:"column:ID;primary_key"`
	Username       string    `gorm:"column:Username;unique;not null"`
	Password       string    `gorm:"column:Password;not null"`
	AuthToken      string    `gorm:"column:AuthToken;not null"`
	Wins           int       `gorm:"column:Wins;default:0"`
	Attempts       int       `gorm:"column:Attempts;default:0"`
	CreatedAt      time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	CorrectGuesses int       `gorm:"column:correct_guesses;default:0"`
}

func (User) TableName() string { return "users" }

type MaxID struct {
	MaxID string `gorm:"max(ID)"`
}

/* ----------------- DB Config ---------------- */

type DBConfig struct {
	DBUser     string `json:"DB_USER"`
	DBPassword string `json:"DB_PASSWORD"`
	DBHost     string `json:"DB_HOST"`
	DBPort     string `json:"DB_PORT"`
	DBName     string `json:"DB_NAME"`
}

var db *gorm.DB
var logger *zap.Logger

/* ----------------- 初始化 ----------------- */

func init() {
	if wd, err := os.Getwd(); err == nil {
		_ = godotenv.Load(filepath.Join(wd, ".env"))
	}
	// 全局 logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

func initDatabase() {
	cc := constant.ClientConfig{
		NamespaceId: os.Getenv("NACOS_NAMESPACE"),
		TimeoutMs:   mustUint(os.Getenv("NACOS_TIMEOUT_MS")),
		Username:    os.Getenv("NACOS_USERNAME"),
		Password:    os.Getenv("NACOS_PASSWORD"),
	}
	sc := []constant.ServerConfig{{
		IpAddr:      os.Getenv("NACOS_SERVER_IP"),
		ContextPath: os.Getenv("NACOS_CONTEXT_PATH"),
		Port:        mustUint(os.Getenv("NACOS_SERVER_PORT")),
	}}

	cfgCli, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc, "clientConfig": cc,
	})
	if err != nil {
		logger.Fatal("create nacos cfg client", zap.Error(err))
	}

	raw, err := cfgCli.GetConfig(vo.ConfigParam{
		DataId: "Prod_DATABASE",
		Group:  "DEFAULT_GROUP",
	})
	if err != nil {
		logger.Fatal("get db config", zap.Error(err))
	}

	var dbc DBConfig
	if err = json.Unmarshal([]byte(raw), &dbc); err != nil {
		logger.Fatal("parse db config", zap.Error(err))
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbc.DBUser, dbc.DBPassword, dbc.DBHost, dbc.DBPort, dbc.DBName)
	db, err = gorm.Open("mysql", dsn)
	if err != nil {
		logger.Fatal("mysql open", zap.Error(err))
	}
	db.AutoMigrate(&User{})
	logger.Info("database connected")
}

/* ----------------- 工具函数 ----------------- */

func mustUint(s string) uint64 {
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		logger.Fatal("parse uint", zap.String("input", s), zap.Error(err))
	}
	return u
}

func closeDatabase() {
	if db != nil {
		_ = db.Close()
	}
}

func getNextUserID() (string, error) {
	var res MaxID
	if err := db.Table("users").Select("MAX(ID) as max_id").Scan(&res).Error; err != nil {
		return "", err
	}
	next := 1
	if res.MaxID != "" {
		cur, err := strconv.Atoi(res.MaxID)
		if err != nil {
			return "", err
		}
		next = cur + 1
	}
	if next > 999999 {
		return "", fmt.Errorf("User ID exceeds 6 digits")
	}
	return fmt.Sprintf("%06d", next), nil
}

func getHealthyInstance(instances []model.Instance) *model.Instance {
	for _, ins := range instances {
		if ins.Healthy {
			return &ins
		}
	}
	return nil
}
