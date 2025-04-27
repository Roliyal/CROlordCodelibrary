package main

import (
	"encoding/json"
	"fmt"
	"os"
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
	"go.uber.org/zap/zapcore"
)

/* ---------- 全局 logger ---------- */

var logger *zap.Logger

func init() {
	_ = godotenv.Load(".env")

	enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		LevelKey:     "severity",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		MessageKey:   "service",
		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	})
	core := zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), zap.InfoLevel)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

/* ---------- 数据模型 ---------- */

type User struct {
	ID        string    `gorm:"column:ID;primary_key"`
	Username  string    `gorm:"column:Username;unique;not null"`
	Password  string    `gorm:"column:Password;not null"`
	AuthToken string    `gorm:"column:AuthToken;not null"`
	Wins      int       `gorm:"column:Wins;default:0"`
	Attempts  int       `gorm:"column:Attempts;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func (User) TableName() string { return "users" }

type dbConf struct {
	DBUser, DBPassword, DBHost, DBPort, DBName string
}

var db *gorm.DB

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
		"serverConfigs": sc, "clientConfig": cc})
	if err != nil {
		logger.Fatal("create config client", zap.Error(err))
	}

	raw, err := cfgCli.GetConfig(vo.ConfigParam{DataId: "Prod_DATABASE", Group: "DEFAULT_GROUP"})
	if err != nil {
		logger.Fatal("get nacos cfg", zap.Error(err))
	}
	var cfg dbConf
	if err = json.Unmarshal([]byte(raw), &cfg); err != nil {
		logger.Fatal("parse db cfg", zap.Error(err))
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err = gorm.Open("mysql", dsn)
	if err != nil {
		logger.Fatal("mysql open", zap.Error(err))
	}
	db.AutoMigrate(&User{})
}

func mustUint(s string) uint64 {
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		logger.Fatal("parseUint", zap.String("input", s), zap.Error(err))
	}
	return u
}

func closeDatabase() {
	if db != nil {
		_ = db.Close()
	}
}

type MaxID struct {
	MaxID string `gorm:"max(ID)"`
}

func getNextUserID() (string, error) {
	var r MaxID
	if err := db.Table("users").Select("MAX(ID) as max_id").Scan(&r).Error; err != nil {
		return "", err
	}
	cur, _ := strconv.Atoi(r.MaxID)
	return fmt.Sprintf("%06d", cur+1), nil
}

func getHealthyInstance(instances []model.Instance) *model.Instance {
	for _, ins := range instances {
		if ins.Healthy {
			return &ins
		}
	}
	return nil
}
