// database.go

package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var db *gorm.DB

// User 结构体
type User struct {
	ID       uint   `gorm:"primary_key"`
	Username string `gorm:"unique"`
	Password string
}

// Game 结构体
type Game struct {
	ID             uint `gorm:"primary_key"`
	TargetNumber   int
	Attempts       int
	CorrectGuesses int
}

// 初始化数据库连接
func initDatabase(dbConfig map[string]string) {
	var err error
	// 处理特殊字符的密码
	password := dbConfig["DB_PASSWORD"]
	// 如果密码中包含特殊字符，建议使用 URL 编码
	// password = url.QueryEscape(dbConfig["DB_PASSWORD"])

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbConfig["DB_USER"],
		password,
		dbConfig["DB_HOST"],
		dbConfig["DB_PORT"],
		dbConfig["DB_NAME"],
	)

	// 打印 DSN，供调试使用（请注意安全性，生产环境下不要打印密码）
	log.Printf("Connecting to database with DSN: %s", dsn)

	db, err = gorm.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}
	// 自动迁移数据库表
	db.AutoMigrate(&User{}, &Game{})
}

// 自定义表名
func (Game) TableName() string {
	return "game"
}

// 获取或创建游戏记录
func getOrCreateGame(user *User) (*Game, error) {
	var game Game
	if err := db.Where("id = ?", user.ID).First(&game).Error; err != nil {
		log.Println("No game record found for user:", user.ID)

		if gorm.IsRecordNotFoundError(err) {
			game.ID = user.ID // 使用 user.ID 作为游戏记录的 ID
			game.TargetNumber = generateTargetNumber()
			game.Attempts = 0
			if err := db.Create(&game).Error; err != nil {
				return nil, err
			}
		} else {
			log.Println("Error querying game record:", err)
			return nil, err
		}
	}
	return &game, nil
}

// 增加尝试次数
func incrementAttempts(game *Game) {
	game.Attempts++
	db.Save(game)
}

// 通过 userID 从 login-service 获取用户信息
func getUserFromUserID(userID uint) (User, error) {
	// 使用 Nacos 发现 login-service
	service, err := NamingClient.GetService(vo.GetServiceParam{
		ServiceName: "login-service",
		GroupName:   "DEFAULT_GROUP",
	})
	if err != nil {
		return User{}, fmt.Errorf("failed to discover login service: %w", err)
	}

	if len(service.Hosts) == 0 {
		log.Println("No instances found for login-service in Nacos")
		return User{}, fmt.Errorf("no healthy login service instance found")
	}

	log.Printf("Found %d instances for login-service in Nacos", len(service.Hosts))
	for i, host := range service.Hosts {
		log.Printf("Instance %d: IP=%s, Port=%d, Healthy=%t", i+1, host.Ip, host.Port, host.Healthy)
	}

	instance := getHealthyInstance(service.Hosts)
	if instance == nil {
		return User{}, fmt.Errorf("no healthy login service instance found")
	}

	// 构建请求 URL
	url := fmt.Sprintf("http://%s:%d/user", instance.Ip, instance.Port)

	// 创建 GET 请求，并设置 X-User-ID 头
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return User{}, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-User-ID", strconv.Itoa(int(userID)))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("error sending request to login service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("login service returned status %d", resp.StatusCode)
	}

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return User{}, fmt.Errorf("error decoding user JSON: %w", err)
	}
	return user, nil
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
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(100) + 1 // 1 到 100
}

// 关闭数据库连接
func closeDatabase() {
	db.Close()
}

// 获取数据库配置从 Nacos
func getDatabaseConfigFromNacos() (map[string]string, error) {
	DataId := "Prod_DATABASE"
	Group := "DEFAULT_GROUP"

	fmt.Printf("Requesting Nacos config with DataId: %s, Group: %s\n", DataId, Group)

	config, err := ConfigClient.GetConfig(vo.ConfigParam{
		DataId: DataId,
		Group:  Group,
	})
	if err != nil {
		return nil, err
	}

	fmt.Printf("Received Nacos config: %s\n", config)

	var dbConfig map[string]string
	err = json.Unmarshal([]byte(config), &dbConfig)
	if err != nil {
		return nil, err
	}

	return dbConfig, nil
}
