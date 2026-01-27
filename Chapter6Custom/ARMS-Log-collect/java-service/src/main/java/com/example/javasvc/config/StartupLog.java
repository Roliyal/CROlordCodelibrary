package com.example.javasvc.config;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.apache.logging.log4j.CloseableThreadContext;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.stereotype.Component;

@Component
public class StartupLog implements ApplicationRunner {
  private static final Logger log = LogManager.getLogger(StartupLog.class);

  private final ArmsLogProperties props;

  public StartupLog(ArmsLogProperties props) {
    this.props = props;
  }

  @Override
  public void run(ApplicationArguments args) {
    try (var ignored = CloseableThreadContext.put("source", "StartupLog")
        .put("category", "startup.config")
        .put("env", props.getEnv())
        .put("version", props.getVersion())
        .put("goHttpBaseUrl", props.getGoHttpBaseUrl())
        .put("goGrpcAddr", props.getGoGrpcAddr())
        .put("httpTimeoutMillis", String.valueOf(props.getHttpTimeoutMillis()))
        .put("grpcTimeoutMillis", String.valueOf(props.getGrpcTimeoutMillis()))) {
      log.info("startup config loaded");
    }
  }
}
