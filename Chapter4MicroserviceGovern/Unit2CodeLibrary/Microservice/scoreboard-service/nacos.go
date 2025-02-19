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

// 全局变量
var NamingClient naming_client.INamingClient
var ConfigClient config_client.IConfigClient

// initNacos 初始化 Nacos 客户端
func initNacos() (naming_client.INamingClient, config_client.IConfigClient, error) {
	// 读取环境变量
	timeoutMs, err := strconv.ParseUint(os.Getenv("NACOS_TIMEOUT_MS"), 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to parse NACOS_TIMEOUT_MS: %v", err)
	}
	nacosPort, err := strconv.ParseUint(os.Getenv("NACOS_SERVER_PORT"), 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to parse NACOS_SERVER_PORT: %v", err)
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         os.Getenv("NACOS_NAMESPACE"),
		TimeoutMs:           timeoutMs,
		NotLoadCacheAtStart: true,
		LogDir:              "logs",
		CacheDir:            "cache",
		Username:            os.Getenv("NACOS_USERNAME"),
		Password:            os.Getenv("NACOS_PASSWORD"),
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
		return nil, nil, fmt.Errorf("Failed to create Nacos naming client: %v", err)
	}
	NamingClient = nc

	// 创建配置客户端
	cc, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create Nacos config client: %v", err)
	}
	ConfigClient = cc

	// 注册服务
	err = registerService(NamingClient, "scoreboard-service", 8085)
	if err != nil {
		return nil, nil, fmt.Errorf("Error registering service: %v", err)
	}

	return nc, cc, nil
}

// getHostIP 获取主机的非回环 IP 地址
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
func registerService(client naming_client.INamingClient, serviceName string, port uint64) error {
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

	log.Printf("Service %s registered successfully at %s:%d", serviceName, hostIP, port)
	return nil
}

// deregisterService 注销服务从 Nacos
func deregisterService(serviceName string, port uint64) error {
	hostIP, err := getHostIP()
	if err != nil {
		return fmt.Errorf("Failed to get host IP address: %w", err)
	}

	_, err = NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          hostIP,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   "DEFAULT_GROUP",
	})
	if err != nil {
		return fmt.Errorf("failed to deregister service instance: %w", err)
	}
	return nil
}
