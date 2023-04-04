package main

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"log"
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

	// New naming client
	nc, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		log.Fatalf("Failed to create Nacos client: %v", err)
	}
	namingClient = nc
}
