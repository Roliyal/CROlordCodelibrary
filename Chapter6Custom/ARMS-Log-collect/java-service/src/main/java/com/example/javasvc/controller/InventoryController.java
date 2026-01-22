package com.example.javasvc.controller;

import com.example.javasvc.grpc.GoBridgeGrpcClient;
import com.example.javasvc.http.GoHttpClient;
import org.apache.logging.log4j.CloseableThreadContext;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/inventory")
public class InventoryController {
  private static final Logger log = LogManager.getLogger(InventoryController.class);

  private final GoHttpClient goHttp;
  private final GoBridgeGrpcClient goGrpc;

  public InventoryController(GoHttpClient goHttp, GoBridgeGrpcClient goGrpc) {
    this.goHttp = goHttp;
    this.goGrpc = goGrpc;
  }

  /** 3/3 Java HTTP endpoint: reserve inventory */
  @PostMapping("/reserve")
  public String reserve(
      @RequestParam(name = "n", defaultValue = "10") int n,
      @RequestHeader(value = "X-Caller-Service", required = false) String caller) {

    try (var ctx =
        CloseableThreadContext.put("source", "InventoryController").put("category", "inventory.reserve")) {

      for (int i = 0; i < Math.max(1, n); i++) {
        log.info("inventory reserve step idx=" + i);
      }

      boolean calledFromGo = caller != null && caller.equalsIgnoreCase("go-service");

      if (!calledFromGo) {
        goHttp.post("/api/payment/refund?n=1", "remote.http.refund");
      } else {
        log.info("skip go-http callback (calledFromGo=true)");
      }

      goGrpc.issueRefund("REFUND", "inventory_reserve");

      log.info("inventory reserved");
      return "OK";
    }
  }
}
