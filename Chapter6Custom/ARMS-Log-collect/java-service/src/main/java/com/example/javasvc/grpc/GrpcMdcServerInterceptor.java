package com.example.javasvc.grpc;

import com.example.javasvc.util.TraceIdUtil;
import io.grpc.*;
import org.apache.logging.log4j.ThreadContext;

import java.util.concurrent.TimeUnit;

/**
 * Production-hardened gRPC inbound interceptor:
 * - Extract/Create traceId from metadata (x-trace-id / traceparent)
 * - Put traceId into io.grpc.Context for propagation
 * - For each listener callback, set MDC fields then clear (prevents leak)
 */
public class GrpcMdcServerInterceptor implements ServerInterceptor {

  private static final Metadata.Key<String> X_TRACE_ID =
      Metadata.Key.of("x-trace-id", Metadata.ASCII_STRING_MARSHALLER);

  private static final Metadata.Key<String> TRACEPARENT =
      Metadata.Key.of("traceparent", Metadata.ASCII_STRING_MARSHALLER);

  @Override
  public <ReqT, RespT> ServerCall.Listener<ReqT> interceptCall(
      ServerCall<ReqT, RespT> call, Metadata headers, ServerCallHandler<ReqT, RespT> next) {

    String traceId = TraceIdUtil.extractOrCreateFromMetadata(headers.get(X_TRACE_ID), headers.get(TRACEPARENT));
    Context ctx = Context.current().withValue(GrpcMdcKeys.TRACE_ID_CTX_KEY, traceId);

    // Wrap call to add response headers too
    ServerCall<ReqT, RespT> forwarding = new ForwardingServerCall.SimpleForwardingServerCall<>(call) {
      @Override
      public void sendHeaders(Metadata responseHeaders) {
        responseHeaders.put(X_TRACE_ID, traceId);
        super.sendHeaders(responseHeaders);
      }
    };

    Context previous = ctx.attach();
    try {
      ServerCall.Listener<ReqT> listener = next.startCall(forwarding, headers);
      return new ForwardingServerCallListener.SimpleForwardingServerCallListener<>(listener) {

        private void mdcOn(String phase) {
          ThreadContext.put("service", "java-service");
          ThreadContext.put("traceId", traceId);
          ThreadContext.put("protocol", "grpc");
          ThreadContext.put("direction", "inbound");
          ThreadContext.put("method", call.getMethodDescriptor().getFullMethodName());
          ThreadContext.put("path", call.getMethodDescriptor().getFullMethodName());
          ThreadContext.put("category", "grpc.inbound." + phase);
          ThreadContext.put("source", "GrpcMdcServerInterceptor");
        }

        private void mdcOff() { ThreadContext.clearAll(); }

        @Override
        public void onMessage(ReqT message) {
          mdcOn("message");
          try { super.onMessage(message); }
          finally { mdcOff(); }
        }

        @Override
        public void onHalfClose() {
          mdcOn("halfclose");
          try { super.onHalfClose(); }
          finally { mdcOff(); }
        }

        @Override
        public void onCancel() {
          mdcOn("cancel");
          try { super.onCancel(); }
          finally { mdcOff(); }
        }

        @Override
        public void onComplete() {
          mdcOn("complete");
          try { super.onComplete(); }
          finally { mdcOff(); }
        }

        @Override
        public void onReady() {
          mdcOn("ready");
          try { super.onReady(); }
          finally { mdcOff(); }
        }
      };
    } finally {
      ctx.detach(previous);
    }
  }
}
