package main

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var configClient *config_client.ConfigClient
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

	cc, err := clients.CreateConfigClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
	})

	if err != nil {
		panic("failed to connect to Nacos")
	}

	configClient = cc.(*config_client.ConfigClient)

	nc, err := clients.CreateNamingClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
	})

	if err != nil {
		panic("failed to create Nacos naming client")
	}

	namingClient = nc.(naming_client.INamingClient)

	registerInstance()
}

func registerInstance() {
	_, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          "127.0.0.1",
		Port:        8084,
		ServiceName: "game-server",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Metadata:    map[string]string{},
		ClusterName: "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	if err != nil {
		panic("failed to register instance with Nacos")
	}
}

func unregisterInstance() {
	_, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          "127.0.0.1",
		Port:        8084,
		ServiceName: "game-server",
		Cluster:     "DEFAULT",
		GroupName:   "DEFAULT_GROUP",
		Ephemeral:   true,
	})

	if err != nil {
		panic("failed to unregister instance with Nacos")
	}
}
