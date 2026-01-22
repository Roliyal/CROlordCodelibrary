package com.example.javasvc.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Configuration
@ConfigurationProperties(prefix = "armslog")
public class ArmsLogProperties {
  private String env = "dev";
  private String version = "1.0.0";
  private String goHttpBaseUrl = "http://localhost:8081";
  private String goGrpcAddr = "localhost:9091";
  private int grpcPort = 9090;

  public String getEnv() { return env; }
  public void setEnv(String env) { this.env = env; }

  public String getVersion() { return version; }
  public void setVersion(String version) { this.version = version; }

  public String getGoHttpBaseUrl() { return goHttpBaseUrl; }
  public void setGoHttpBaseUrl(String goHttpBaseUrl) { this.goHttpBaseUrl = goHttpBaseUrl; }

  public String getGoGrpcAddr() { return goGrpcAddr; }
  public void setGoGrpcAddr(String goGrpcAddr) { this.goGrpcAddr = goGrpcAddr; }

  public int getGrpcPort() { return grpcPort; }
  public void setGrpcPort(int grpcPort) { this.grpcPort = grpcPort; }
}
