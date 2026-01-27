package com.example.javasvc.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "armslog")
public class ArmsLogProperties {
  private String env = "dev";
  private String version = "1.0.0";

  /** Base URL of go-service HTTP endpoints, e.g. http://go-service:8081 */
  private String goHttpBaseUrl = "http://localhost:8081";

  /** Address of go-service gRPC server, e.g. go-service:9091 */
  private String goGrpcAddr = "localhost:9091";

  /** Port for this java-service's inbound gRPC server. */
  private int grpcPort = 9090;

  /** HTTP client timeout (connect + read) in milliseconds. */
  private long httpTimeoutMillis = 1500;

  /** gRPC client per-call deadline in milliseconds. */
  private long grpcTimeoutMillis = 2000;

  public String getEnv() {
    return env;
  }

  public void setEnv(String env) {
    this.env = env;
  }

  public String getVersion() {
    return version;
  }

  public void setVersion(String version) {
    this.version = version;
  }

  public String getGoHttpBaseUrl() {
    return goHttpBaseUrl;
  }

  public void setGoHttpBaseUrl(String goHttpBaseUrl) {
    this.goHttpBaseUrl = goHttpBaseUrl;
  }

  public String getGoGrpcAddr() {
    return goGrpcAddr;
  }

  public void setGoGrpcAddr(String goGrpcAddr) {
    this.goGrpcAddr = goGrpcAddr;
  }

  public int getGrpcPort() {
    return grpcPort;
  }

  public void setGrpcPort(int grpcPort) {
    this.grpcPort = grpcPort;
  }

  public long getHttpTimeoutMillis() {
    return httpTimeoutMillis;
  }

  public void setHttpTimeoutMillis(long httpTimeoutMillis) {
    this.httpTimeoutMillis = httpTimeoutMillis;
  }

  public long getGrpcTimeoutMillis() {
    return grpcTimeoutMillis;
  }

  public void setGrpcTimeoutMillis(long grpcTimeoutMillis) {
    this.grpcTimeoutMillis = grpcTimeoutMillis;
  }
}
