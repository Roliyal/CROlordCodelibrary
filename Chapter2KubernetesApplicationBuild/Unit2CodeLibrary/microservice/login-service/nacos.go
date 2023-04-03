package main

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
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

	iClient, err := clients.CreateConfigClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
		constant.KEY_USERNAME:       "nacos",
		constant.KEY_PASSWORD:       "nacos",
	})

	if err != nil {
		panic("failed to connect to Nacos")
	}

	configClient = iClient.(*config_client.ConfigClient)

	nClient, err := clients.CreateNamingClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
	})

	if err != nil {
		panic("failed to connect to Nacos")
	}

	namingClient = nClient
}
