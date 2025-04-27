// nacos.go
package main

import (
	"fmt"
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
	serverConfigs := []constant.ServerConfig{{
		IpAddr:      os.Getenv("NACOS_SERVER_IP"),
		ContextPath: os.Getenv("NACOS_CONTEXT_PATH"),
		Port:        nacosPort,
	}}

	nc, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create Nacos naming client: %v", err)
	}
	NamingClient = nc

	cc, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create Nacos config client: %v", err)
	}
	ConfigClient = cc

	if err = registerService(NamingClient, "scoreboard-service", 8085); err != nil {
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
		if ip, _, _ := net.ParseCIDR(addr.String()); ip != nil && !ip.IsLoopback() && ip.To4() != nil {
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

	ok, err := client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          hostIP,
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
	if !ok {
		return fmt.Errorf("Failed to register service")
	}
	zapLog.Infow("Service registered", "service", serviceName, "ip", hostIP, "port", port)
	return nil
}

// deregisterService 注销服务从 Nacos
func deregisterService(serviceName string, port uint64) error {
	hostIP, err := getHostIP()
	if err != nil {
		return fmt.Errorf("Failed to get host IP address: %w", err)
	}
	if _, err = NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          hostIP,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   "DEFAULT_GROUP",
	}); err != nil {
		return fmt.Errorf("failed to deregister service instance: %w", err)
	}
	zapLog.Infow("Service deregistered", "service", serviceName)
	return nil
}
