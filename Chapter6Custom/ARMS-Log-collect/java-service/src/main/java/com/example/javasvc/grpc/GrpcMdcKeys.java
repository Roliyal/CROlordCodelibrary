package com.example.javasvc.grpc;

import io.grpc.Context;

public final class GrpcMdcKeys {
  private GrpcMdcKeys() {}
  public static final Context.Key<String> TRACE_ID_CTX_KEY = Context.key("traceId");
}
