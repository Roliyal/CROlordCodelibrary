package main

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var namingClient naming_client.INamingClient

func initNacos() {
	clientConfig := constant.ClientConfig{
		NamespaceId: "public",
		TimeoutMs:   5000,
		Username:    "nacos",
		Password:    "nacos",
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "mse-c00253114-p.nacos-ans.mse.aliyuncs.com",
			ContextPath: "/nacos",
			Port:        8848,
		},
	}

	nc, err := clients.CreateNamingClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
	})
	if err != nil {
		panic("failed to create Nacos client")
	}
	namingClient = nc

	// Register the service instance
	success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          "127.0.0.1",
		Port:        8084,
		ServiceName: "game-service",
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Metadata:    map[string]string{"version": "1.0"},
	})
	if err != nil || !success {
		panic("failed to register service instance")
	}
}

func getLoginServiceURL() string {
	// Discover the login service using Nacos
	service, err := namingClient.GetService(vo.GetServiceParam{
		ServiceName: "login-service", // 使用正确的服务名称
		GroupName:   "DEFAULT_GROUP",
	})
	if err != nil {
		panic("failed to discover login service")
	}

	// Choose the first instance for now
	instance := service.Hosts[0]
	url := fmt.Sprintf("http://%s:%d", instance.Ip, instance.Port)
	fmt.Printf("Login service URL: %s\n", url) // 添加这行代码
	return url
}
