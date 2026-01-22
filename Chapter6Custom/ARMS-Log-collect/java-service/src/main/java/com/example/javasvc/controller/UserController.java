package com.example.javasvc.controller;

import com.example.javasvc.grpc.GoBridgeGrpcClient;
import com.example.javasvc.http.GoHttpClient;
import org.apache.logging.log4j.CloseableThreadContext;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/user")
public class UserController {
  private static final Logger log = LogManager.getLogger(UserController.class);

  private final GoHttpClient goHttp;
  private final GoBridgeGrpcClient goGrpc;

  public UserController(GoHttpClient goHttp, GoBridgeGrpcClient goGrpc) {
    this.goHttp = goHttp;
    this.goGrpc = goGrpc;
  }

  /** 2/3 Java HTTP endpoint: get user */
  @GetMapping("/get")
  public String get(
      @RequestParam(name = "n", defaultValue = "10") int n,
      @RequestHeader(value = "X-Caller-Service", required = false) String caller) {

    try (var ctx =
        CloseableThreadContext.put("source", "UserController").put("category", "user.get")) {

      for (int i = 0; i < Math.max(1, n); i++) {
        log.info("user get step idx=" + i);
      }

      boolean calledFromGo = caller != null && caller.equalsIgnoreCase("go-service");

      // ✅ 断环：如果是 go-service 调进来的，就不要再 HTTP 回调 go-service
      if (!calledFromGo) {
        goHttp.post("/api/payment/query?n=1", "remote.http.query");
      } else {
        log.info("skip go-http callback (calledFromGo=true)");
      }

      // gRPC 调 Go 不会再触发 Java HTTP，因此安全
      goGrpc.queryPayment("QUERY", "user_get");

      log.info("user got");
      return "OK";
    }
  }
}
