package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"os"
	"strconv"
)

func initNacos() (naming_client.INamingClient, config_client.IConfigClient, error) {
	nacosNamingClient, err := createNacosNamingClient()
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating Nacos naming client: %v", err)
	}

	nacosConfigClient, err := createNacosConfigClient()
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating Nacos config client: %v", err)
	}

	err = registerService(nacosNamingClient, "scoreboard-service", "localhost", 8085)
	if err != nil {
		return nil, nil, fmt.Errorf("Error registering service: %v", err)
	}

	return nacosNamingClient, nacosConfigClient, nil
}

func mustParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func createNacosConfigClient() (config_client.IConfigClient, error) {
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: os.Getenv("NACOS_SERVER_IP"),
			Port:   uint64(mustParseInt(os.Getenv("NACOS_SERVER_PORT"))),
		},
	}

	timeoutMs, _ := strconv.Atoi(os.Getenv("NACOS_TIMEOUT_MS"))

	clientConfig := constant.ClientConfig{
		NamespaceId:         os.Getenv("NACOS_NAMESPACE"),
		TimeoutMs:           uint64(timeoutMs),
		Username:            os.Getenv("NACOS_USERNAME"),
		Password:            os.Getenv("NACOS_PASSWORD"),
		LogDir:              "nacos-log",
		CacheDir:            "nacos-cache",
		UpdateThreadNum:     2,
		NotLoadCacheAtStart: true,
	}

	nacosConfigClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})

	return nacosConfigClient, err
}
func createNacosNamingClient() (naming_client.INamingClient, error) {
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: os.Getenv("NACOS_SERVER_IP"),
			Port:   uint64(mustParseInt(os.Getenv("NACOS_SERVER_PORT"))),
		},
	}

	timeoutMs, _ := strconv.Atoi(os.Getenv("NACOS_TIMEOUT_MS"))

	clientConfig := constant.ClientConfig{
		NamespaceId:         os.Getenv("NACOS_NAMESPACE"),
		TimeoutMs:           uint64(timeoutMs),
		Username:            os.Getenv("NACOS_USERNAME"),
		Password:            os.Getenv("NACOS_PASSWORD"),
		LogDir:              "nacos-log",
		CacheDir:            "nacos-cache",
		UpdateThreadNum:     2,
		NotLoadCacheAtStart: true,
	}

	nacosNamingClient, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})

	return nacosNamingClient, err
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
