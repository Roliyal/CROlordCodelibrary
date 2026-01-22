package com.example.javasvc.grpc;

import com.example.bridge.v1.JavaBridgeGrpc;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.ServerInterceptors;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import com.example.javasvc.config.ArmsLogProperties;

@Component
public class JavaGrpcServerRunner implements CommandLineRunner {

  private static final Logger log = LogManager.getLogger(JavaGrpcServerRunner.class);
  private final ArmsLogProperties props;
  private Server server;

  public JavaGrpcServerRunner(ArmsLogProperties props) {
    this.props = props;
  }

  @Override
  public void run(String... args) throws Exception {
    server = ServerBuilder.forPort(props.getGrpcPort())
        .addService(ServerInterceptors.intercept(new JavaBridgeService(), new GrpcMdcServerInterceptor()))
        .build()
        .start();

    log.info("java grpc server started on port " + props.getGrpcPort());

    Runtime.getRuntime().addShutdownHook(new Thread(() -> {
      log.info("java grpc server shutting down");
      if (server != null) server.shutdown();
    }));
  }
}
