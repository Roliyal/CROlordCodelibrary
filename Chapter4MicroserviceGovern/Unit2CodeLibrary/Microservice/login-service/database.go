// database.go

package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	"math/rand"
)

// User 结构体
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

// TableName 显式指定表名为 `users`
func (User) TableName() string {
	return "users"
}

// DBConfig 结构体
type DBConfig struct {
	DBUser     string `json:"DB_USER"`
	DBPassword string `json:"DB_PASSWORD"`
	DBHost     string `json:"DB_HOST"`
	DBPort     string `json:"DB_PORT"`
	DBName     string `json:"DB_NAME"`
}

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

	rand.Seed(time.Now().UnixNano()) // 初始化随机数种子
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

	// 创建配置客户端
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

	dbConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbConfig.DBUser,
		dbConfig.DBPassword,
		dbConfig.DBHost,
		dbConfig.DBPort,
		dbConfig.DBName,
	)
	log.Printf("Connecting to database with DSN: %s", dbConnectionString)

	db, err = gorm.Open("mysql", dbConnectionString)
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		panic("failed to connect to database")
	}

	// 自动迁移数据库表
	db.AutoMigrate(&User{})
}

// MaxID 结构体用于接收 MAX(ID) 的查询结果
type MaxID struct {
	MaxID string `gorm:"max(id)"`
}

// getNextUserID 生成下一个唯一的6位数用户ID
func getNextUserID() (string, error) {
	var result MaxID
	// 查询当前最大的ID
	err := db.Model(&User{}).Select("MAX(ID) as max_id").Scan(&result).Error
	if err != nil {
		return "", err
	}

	var nextID int
	if result.MaxID == "" {
		nextID = 1
	} else {
		currentID, err := strconv.Atoi(result.MaxID)
		if err != nil {
			return "", err
		}
		nextID = currentID + 1
	}

	// 生成6位数ID，补零
	if nextID > 999999 {
		return "", fmt.Errorf("User ID exceeds 6 digits")
	}
	return fmt.Sprintf("%06d", nextID), nil
}

func mustParseUint(s string) uint64 {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse uint from string: %v", err)
	}
	return i
}

func closeDatabase() {
	if db != nil {
		db.Close()
	}
}

// 从实例列表中获取第一个健康的实例
func getHealthyInstance(instances []model.Instance) *model.Instance {
	for _, instance := range instances {
		if instance.Healthy {
			return &instance
		}
	}
	return nil
}

// 生成 1 到 100 之间的随机数
func generateTargetNumber() int {
	rand.Seed(time.Now().UnixNano()) // 使用 math/rand 包
	return rand.Intn(100) + 1        // 1 到 100
}
