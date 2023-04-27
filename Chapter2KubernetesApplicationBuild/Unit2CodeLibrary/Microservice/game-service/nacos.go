package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"net"
	"os"
	"strconv"
)

var NamingClient naming_client.INamingClient
var ConfigClient config_client.IConfigClient

func initNacos() {
	// 读取.env文件
	err := godotenv.Load(".env")
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
	fmt.Printf("Nacos server config: %v\n", serverConfigs) // 输出 Nacos 服务器配置

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

func subscribeLoginService() {
	err := NamingClient.Subscribe(&vo.SubscribeParam{
		ServiceName: "login-service",
		GroupName:   "DEFAULT_GROUP",
		Clusters:    []string{"DEFAULT"},
		SubscribeCallback: func(services []model.SubscribeService, err error) {
			if err != nil {
				fmt.Printf("Error in SubscribeCallback: %v\n", err)
				return
			}

			fmt.Println("Login service instances update:") // 输出订阅成功信息
			for _, service := range services {
				fmt.Printf("Instance: IP=%s, Port=%d\n", service.Ip, service.Port)
			}
		},
	})
	if err != nil {
		panic("failed to subscribe to login-service")
	} else {
		fmt.Println("Successfully subscribed to login-service") // 输出订阅成功信息
	}
}

func parseInt(value string, defaultValue int) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return result
}
func getHostIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}
		if !ip.IsLoopback() && ip.To4() != nil {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("No valid IP address found")
}

func registerService(client naming_client.INamingClient, serviceName, ip string, port uint64) error {
	hostIP, err := getHostIP()
	if err != nil {
		return fmt.Errorf("Failed to get host IP address: %w", err)
	}

	success, err := client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          hostIP, // 使用动态获取的宿主机 IP 地址
		Port:        port,
		ServiceName: serviceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	})

	if err != nil {
		return fmt.Errorf("registerService error: %w", err)
	}

	if !success {
		return fmt.Errorf("Failed to register service")
	}

	return nil
}

func deregisterGameService() {
	hostIP, err := getHostIP()

	_, err = NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{

		Ip:          hostIP,
		Port:        8084,
		ServiceName: "game-service",
		GroupName:   "DEFAULT_GROUP",
	})
	if err != nil {
		panic("failed to deregister game service instance")
	}
}
