package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"os"
	"strconv"
)

var NamingClient naming_client.INamingClient
var ConfigClient config_client.IConfigClient

func initNacos() {
	// 读取.env文件
	err := godotenv.Load("../.env")
	if err != nil {
		panic("Error loading .env file")
	}

	clientConfig := constant.ClientConfig{
		NamespaceId: os.Getenv("NACOS_NAMESPACE"),
		TimeoutMs:   uint64(parseInt(os.Getenv("NACOS_TIMEOUT_MS"), 5000)),
		Username:    os.Getenv("NACOS_USERNAME"),
		Password:    os.Getenv("NACOS_PASSWORD"),
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      os.Getenv("NACOS_SERVER_IP"),
			ContextPath: os.Getenv("NACOS_CONTEXT_PATH"),
			Port:        uint64(parseInt(os.Getenv("NACOS_SERVER_PORT"), 8848)),
		},
	}

	nc, err := clients.CreateNamingClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
	})
	if err != nil {
		panic("failed to create Nacos naming client")
	}
	NamingClient = nc

	// 创建Nacos配置客户端
	cc, err := clients.CreateConfigClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
	})
	if err != nil {
		panic("failed to create Nacos config client")
	}
	ConfigClient = cc
}

func getDatabaseConfig() (string, error) {
	content, err := ConfigClient.GetConfig(vo.ConfigParam{
		DataId: "Prod_DATABASE",
		Group:  "DEFAULT_GROUP",
	})
	if err != nil {
		return "", err
	}
	return content, nil
}

func parseInt(value string, defaultValue int) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return result
}

func getLoginServiceURL() string {
	// Discover the login service using Nacos
	service, err := NamingClient.GetService(vo.GetServiceParam{
		ServiceName: "login-service",
		GroupName:   "DEFAULT_GROUP",
	})
	if err != nil {
		panic("failed to discover login service")
	}

	// Choose the first instance for now
	instance := service.Hosts[0]
	url := fmt.Sprintf("http://%s:%d", instance.Ip, instance.Port)
	fmt.Printf("Login service URL: %s\n", url) //
	return url
}
