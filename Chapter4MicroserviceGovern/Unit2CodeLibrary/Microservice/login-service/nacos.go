package main

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
	"net"
	"os"
)

var NamingClient naming_client.INamingClient
var ConfigClient config_client.IConfigClient

// ---------- 初始化 ----------
func initNacos() {
	cc := constant.ClientConfig{
		NamespaceId: os.Getenv("NACOS_NAMESPACE"),
		TimeoutMs:   mustUint(os.Getenv("NACOS_TIMEOUT_MS")),
		Username:    os.Getenv("NACOS_USERNAME"),
		Password:    os.Getenv("NACOS_PASSWORD"),
	}
	sc := []constant.ServerConfig{{
		IpAddr:      os.Getenv("NACOS_SERVER_IP"),
		ContextPath: os.Getenv("NACOS_CONTEXT_PATH"),
		Port:        mustUint(os.Getenv("NACOS_SERVER_PORT")),
	}}

	var err error
	NamingClient, err = clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": sc, "clientConfig": cc,
	})
	if err != nil {
		log.Fatalf("create naming client: %v", err)
	}
	ConfigClient, err = clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc, "clientConfig": cc,
	})
	if err != nil {
		log.Fatalf("create config client: %v", err)
	}
}

// ---------- 工具 ----------

func getHostIP() (string, error) {
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no IP found")
}

func registerService(c naming_client.INamingClient, name, ip string, port uint64) error {
	ok, err := c.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: name,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("register failed")
	}
	return nil
}

func deregisterLoginService() {
	hostIP, _ := getHostIP()
	if _, err := NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          hostIP,
		Port:        8083,
		ServiceName: "login-service",
	}); err != nil {
		log.Printf("deregister error: %v", err)
	}
}
