package com.example.javasvc.controller;

import com.example.javasvc.grpc.GoBridgeGrpcClient;
import com.example.javasvc.http.GoHttpClient;
import org.apache.logging.log4j.CloseableThreadContext;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/order")
public class OrderController {
  private static final Logger log = LogManager.getLogger(OrderController.class);

  private final GoHttpClient goHttp;
  private final GoBridgeGrpcClient goGrpc;

  public OrderController(GoHttpClient goHttp, GoBridgeGrpcClient goGrpc) {
    this.goHttp = goHttp;
    this.goGrpc = goGrpc;
  }

  /**
   * 1/3 Java HTTP endpoint: create order
   * - emits burst logs (n)
   * - calls Go via HTTP + gRPC
   */
  @PostMapping("/create")
  public String create(@RequestParam(name = "n", defaultValue = "10") int n) {
    try (var ctx = CloseableThreadContext.put("source", "OrderController").put("category", "order.create")) {
      for (int i = 0; i < Math.max(1, n); i++) {
        log.info("order create step idx=" + i);
      }

      goHttp.post("/api/payment/pay?n=5", "remote.http.pay");
      goGrpc.processPayment("PAY", "order_create");

      log.info("order created");
      return "OK";
    }
  }
}
