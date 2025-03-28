package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var NamingClient naming_client.INamingClient
var ConfigClient config_client.IConfigClient

func initNacos() {
	timeoutMs, err := strconv.ParseUint(os.Getenv("NACOS_TIMEOUT_MS"), 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse NACOS_TIMEOUT_MS: %v", err)
	}
	nacosPort, err := strconv.ParseUint(os.Getenv("NACOS_SERVER_PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse NACOS_SERVER_PORT: %v", err)
	}

	clientConfig := constant.ClientConfig{
		NamespaceId: os.Getenv("NACOS_NAMESPACE"),
		TimeoutMs:   timeoutMs,
		Username:    os.Getenv("NACOS_USERNAME"),
		Password:    os.Getenv("NACOS_PASSWORD"),
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      os.Getenv("NACOS_SERVER_IP"),
			ContextPath: os.Getenv("NACOS_CONTEXT_PATH"),
			Port:        nacosPort,
		},
	}

	// 创建命名客户端
	nc, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		log.Fatalf("Failed to create Nacos naming client: %v", err)
	}
	NamingClient = nc

	// 创建配置客户端
	cc, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		log.Fatalf("Failed to create Nacos config client: %v", err)
	}
	ConfigClient = cc
}

// getHostIP 获取非回环的 IPv4 地址
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

// registerService 注册服务到 Nacos
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

// deregisterLoginService 从 Nacos 注销服务
func deregisterLoginService() {
	hostIP, err := getHostIP()
	if err != nil {
		log.Printf("Failed to get host IP for deregistration: %v\n", err)
		return
	}

	_, err = NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          hostIP,
		Port:        8083,
		ServiceName: "login-service",
		GroupName:   "DEFAULT_GROUP",
	})
	if err != nil {
		log.Printf("Error deregistering login service instance: %v\n", err)
	} else {
		log.Println("Login service deregistered successfully")
	}
}
