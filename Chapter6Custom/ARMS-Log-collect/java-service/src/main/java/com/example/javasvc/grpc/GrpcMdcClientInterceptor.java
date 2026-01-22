package com.example.javasvc.grpc;

import io.grpc.*;
import org.apache.logging.log4j.ThreadContext;

/**
 * Production-hardened gRPC outbound interceptor:
 * - attach x-trace-id and traceparent from MDC
 */
public class GrpcMdcClientInterceptor implements ClientInterceptor {

  private static final Metadata.Key<String> X_TRACE_ID =
      Metadata.Key.of("x-trace-id", Metadata.ASCII_STRING_MARSHALLER);

  private static final Metadata.Key<String> TRACEPARENT =
      Metadata.Key.of("traceparent", Metadata.ASCII_STRING_MARSHALLER);

  @Override
  public <ReqT, RespT> ClientCall<ReqT, RespT> interceptCall(
      MethodDescriptor<ReqT, RespT> method, CallOptions callOptions, Channel next) {

    return new ForwardingClientCall.SimpleForwardingClientCall<>(next.newCall(method, callOptions)) {
      @Override
      public void start(Listener<RespT> responseListener, Metadata headers) {
        String traceId = ThreadContext.get("traceId");
        if (traceId != null && !traceId.isBlank()) {
          headers.put(X_TRACE_ID, traceId);
          headers.put(TRACEPARENT, "00-" + traceId + "-0000000000000000-01");
        }
        super.start(responseListener, headers);
      }
    };
  }
}
