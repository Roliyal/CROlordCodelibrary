package com.example.javasvc.grpc;

import com.example.bridge.v1.ActionReply;
import com.example.bridge.v1.ActionRequest;
import com.example.bridge.v1.JavaBridgeGrpc;
import io.grpc.Context;
import io.grpc.stub.StreamObserver;
import org.apache.logging.log4j.CloseableThreadContext;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.apache.logging.log4j.ThreadContext;

import java.util.concurrent.ThreadLocalRandom;

/**
 * Java gRPC server: 3 methods, emits INFO/WARN/ERROR structured logs.
 */
public class JavaBridgeService extends JavaBridgeGrpc.JavaBridgeImplBase {
  private static final Logger log = LogManager.getLogger(JavaBridgeService.class);

  @Override
  public void validateUser(ActionRequest request, StreamObserver<ActionReply> responseObserver) {
    handle("grpc.java.validate_user", "JavaBridgeService", "ValidateUser", request, responseObserver, "USER_OK");
  }

  @Override
  public void reserveInventory(ActionRequest request, StreamObserver<ActionReply> responseObserver) {
    handle("grpc.java.reserve_inventory", "JavaBridgeService", "ReserveInventory", request, responseObserver, "INV_RESERVED");
  }

  @Override
  public void auditOrder(ActionRequest request, StreamObserver<ActionReply> responseObserver) {
    handle("grpc.java.audit_order", "JavaBridgeService", "AuditOrder", request, responseObserver, "AUDIT_OK");
  }

  private void handle(String category, String source, String method,
                      ActionRequest req, StreamObserver<ActionReply> obs, String result) {

    long start = System.nanoTime();
    String traceId = ThreadContext.get("traceId");
    if (traceId == null || traceId.isBlank()) {
      traceId = GrpcMdcKeys.TRACE_ID_CTX_KEY.get();
      if (traceId != null) ThreadContext.put("traceId", traceId);
    }

    try (var ctx = CloseableThreadContext.put("category", category)
        .put("source", source)
        .put("protocol", "grpc")
        .put("direction", "inbound")
        .put("method", method)
        .put("path", method)) {

      log.info("grpc request received action=" + req.getAction());

      // business logs (non-debug)
      for (int i = 0; i < 5; i++) {
        log.info("business check step=" + i);
      }
      if (ThreadLocalRandom.current().nextInt(100) < 5) {
        log.warn("business warning: downstream latency high");
      }

      long costMs = (System.nanoTime() - start) / 1_000_000;
      try (var ctx2 = CloseableThreadContext.put("costMs", String.valueOf(costMs)).put("status", "OK")) {
        log.info("grpc request handled");
      }

      obs.onNext(ActionReply.newBuilder()
          .setTraceId(req.getTraceId())
          .setCode(0)
          .setResult(result)
          .build());
      obs.onCompleted();

    } catch (Exception e) {
      long costMs = (System.nanoTime() - start) / 1_000_000;
      try (var ctx = CloseableThreadContext.put("category", "grpc.java.error")
          .put("source", source)
          .put("costMs", String.valueOf(costMs))
          .put("status", "ERR")
          .put("errorType", e.getClass().getName())
          .put("errorMessage", safe(e.getMessage()))
          .put("errorStack", e.toString())) {
        log.error("grpc handler error");
      }
      obs.onError(e);
    }
  }

  private static String safe(String s) { return s == null ? "" : s; }
}
