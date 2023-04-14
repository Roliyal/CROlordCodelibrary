package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var db *gorm.DB

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could not determine working directory: %v", err)
	}
	envPath := filepath.Join(pwd, ".env")
	err = godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading .env file from %s: %v", envPath, err)
	}
}

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
type DBConfig struct {
	DBUser     string `json:"DB_USER"`
	DBPassword string `json:"DB_PASSWORD"`
	DBHost     string `json:"DB_HOST"`
	DBPort     string `json:"DB_PORT"`
	DBName     string `json:"DB_NAME"`
}

func initDatabase() {
	clientConfig := constant.ClientConfig{
		NamespaceId: os.Getenv("NACOS_NAMESPACE"),
		TimeoutMs:   mustParseUint(os.Getenv("NACOS_TIMEOUT_MS")),
		Username:    os.Getenv("NACOS_USERNAME"),
		Password:    os.Getenv("NACOS_PASSWORD"),
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      os.Getenv("NACOS_SERVER_IP"),
			ContextPath: os.Getenv("NACOS_CONTEXT_PATH"),
			Port:        mustParseUint(os.Getenv("NACOS_SERVER_PORT")),
		},
	}

	// 获取 Nacos 配置
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})

	if err != nil {
		log.Fatalf("Failed to create Nacos config client: %v", err)
	}

	// 获取 Nacos 中的数据库配置
	dataId := "Prod_DATABASE" // 请替换为您在 Nacos 中设置的数据 ID
	group := "DEFAULT_GROUP"  // 请替换为您在 Nacos 中设置的组
	dbConfigContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})

	if err != nil {
		log.Fatalf("Failed to get database config from Nacos: %v", err)
	}

	// 解析 JSON 配置
	var dbConfig DBConfig
	err = json.Unmarshal([]byte(dbConfigContent), &dbConfig)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	dbConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBHost, dbConfig.DBPort, dbConfig.DBName)
	db, err = gorm.Open("mysql", dbConnectionString)

	fmt.Println("The connection string is:", dbConnectionString)

	db, err = gorm.Open("mysql", dbConnectionString)
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		panic("failed to connect to database")
	}

	// 创建数据库表
	db.Table("user").AutoMigrate(&User{})
	// 设置AuthToken的默认值
}

func mustParseUint(s string) uint64 {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse uint from string: %v", err)
	}
	return i
}

func closeDatabase() {
	db.Close()
}
