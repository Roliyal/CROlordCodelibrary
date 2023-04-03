package main

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func createNacosClient() (naming_client.INamingClient, error) {
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: "mse-c00253114-p.nacos-ans.mse.aliyuncs.com",
			Port:   80,
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           5000,
		LogDir:              "nacos-log",
		CacheDir:            "nacos-cache",
		UpdateThreadNum:     2,
		NotLoadCacheAtStart: true,
	}

	nacosClient, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})

	return nacosClient, err
}

func registerService(client naming_client.INamingClient, serviceName, ip string, port uint64) error {
	success, err := client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	})

	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("Failed to register service")
	}

	return nil
}

func deregisterService(client naming_client.INamingClient, serviceName, ip string, port uint64) error {
	success, err := client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		Ephemeral:   true,
	})

	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("Failed to deregister service")
	}

	return nil
}
