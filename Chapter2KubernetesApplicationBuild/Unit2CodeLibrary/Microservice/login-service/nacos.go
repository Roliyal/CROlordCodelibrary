package main

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"log"
	"os"
	"strconv"
)

var namingClient naming_client.INamingClient

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
