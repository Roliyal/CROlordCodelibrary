package com.example.javasvc.grpc;

import com.example.bridge.v1.ActionReply;
import com.example.bridge.v1.ActionRequest;
import com.example.bridge.v1.GoBridgeGrpc;
import com.example.javasvc.config.ArmsLogProperties;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.ClientInterceptors;
import org.apache.logging.log4j.CloseableThreadContext;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.stereotype.Component;

import java.time.Duration;
import java.util.concurrent.TimeUnit;

/**
 * Production: reuse channel + timeouts + structured outbound logs.
 */
@Component
public class GoBridgeGrpcClient {
  private static final Logger log = LogManager.getLogger(GoBridgeGrpcClient.class);

  private final ArmsLogProperties props;
  private final ManagedChannel channel;
  private final GoBridgeGrpc.GoBridgeBlockingStub stub;

  public GoBridgeGrpcClient(ArmsLogProperties props) {
    this.props = props;
    String addr = props.getGoGrpcAddr();
    String host = addr.contains(":") ? addr.split(":")[0] : addr;
    int port = addr.contains(":") ? Integer.parseInt(addr.split(":")[1]) : 9091;

    this.channel = ManagedChannelBuilder.forAddress(host, port)
        .usePlaintext()
        .build();

    this.stub = GoBridgeGrpc.newBlockingStub(ClientInterceptors.intercept(channel, new GrpcMdcClientInterceptor()));
  }

  public ActionReply processPayment(String action, String payload) {
    return call("remote.grpc.process_payment", "ProcessPayment", action, payload, (s, req) -> s.processPayment(req));
  }

  public ActionReply issueRefund(String action, String payload) {
    return call("remote.grpc.issue_refund", "IssueRefund", action, payload, (s, req) -> s.issueRefund(req));
  }

  public ActionReply queryPayment(String action, String payload) {
    return call("remote.grpc.query_payment", "QueryPayment", action, payload, (s, req) -> s.queryPayment(req));
  }

  private interface RpcInvoker {
    ActionReply invoke(GoBridgeGrpc.GoBridgeBlockingStub s, ActionRequest req);
  }

  private ActionReply call(String category, String method, String action, String payload, RpcInvoker invoker) {
    long start = System.nanoTime();
    try (var ctx = CloseableThreadContext.put("source", "GoBridgeGrpcClient")
        .put("category", category)
        .put("protocol", "grpc")
        .put("direction", "outbound")
        .put("method", method)
        .put("path", method)
        .put("remoteService", "go-service")) {

      log.info("remote grpc call start");
      String traceId = org.apache.logging.log4j.ThreadContext.get("traceId");
      if (traceId == null) traceId = "";
      ActionRequest req = ActionRequest.newBuilder()
          .setTraceId(traceId)
          .setAction(action)
          .setPayload(payload)
          .build();

      ActionReply reply = invoker.invoke(
          stub.withDeadlineAfter(props.getGrpcTimeoutMillis(), TimeUnit.MILLISECONDS),
          req);

      long costMs = (System.nanoTime() - start) / 1_000_000;
      try (var ctx2 = CloseableThreadContext.put("costMs", String.valueOf(costMs)).put("status", "OK")) {
        log.info("remote grpc call done");
      }
      return reply;

    } catch (Exception e) {
      long costMs = (System.nanoTime() - start) / 1_000_000;
      try (var ctx = CloseableThreadContext.put("source", "GoBridgeGrpcClient")
          .put("category", "remote.grpc.error")
          .put("protocol", "grpc")
          .put("direction", "outbound")
          .put("method", method)
          .put("path", method)
          .put("remoteService", "go-service")
          .put("costMs", String.valueOf(costMs))
          .put("status", "ERR")
          .put("errorType", e.getClass().getName())
          .put("errorMessage", safe(e.getMessage()))
          .put("errorStack", e.toString())) {
        log.error("remote grpc call failed");
      }
      throw e;
    }
  }

  private static String safe(String s) { return s == null ? "" : s; }
}
