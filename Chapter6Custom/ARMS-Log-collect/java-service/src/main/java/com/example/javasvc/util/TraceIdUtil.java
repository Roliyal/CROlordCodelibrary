package com.example.javasvc.util;

import jakarta.servlet.http.HttpServletRequest;

import java.util.Locale;
import java.util.Optional;
import java.util.UUID;

public final class TraceIdUtil {
  private TraceIdUtil() {}

  public static String extractOrCreateHttp(HttpServletRequest r) {
    String x = safe(r.getHeader("X-Trace-Id"));
    if (!x.isBlank()) return x;

    String tp = safe(r.getHeader("traceparent"));
    String fromTp = parseTraceparent(tp);
    if (!fromTp.isBlank()) return fromTp;

    return newTraceId32();
  }

  public static String extractOrCreateFromMetadata(String xTraceId, String traceparent) {
    String x = safe(xTraceId);
    if (!x.isBlank()) return x;

    String fromTp = parseTraceparent(safe(traceparent));
    if (!fromTp.isBlank()) return fromTp;

    return newTraceId32();
  }

  /** W3C traceparent: version-traceid-spanid-flags */
  static String parseTraceparent(String tp) {
    if (tp == null) return "";
    String[] parts = tp.trim().split("-");
    if (parts.length != 4) return "";
    String traceId = parts[1].toLowerCase(Locale.ROOT);
    if (traceId.length() != 32) return "";
    for (char c : traceId.toCharArray()) {
      if (!((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'))) return "";
    }
    // invalid all-zero traceId not allowed
    if (traceId.chars().allMatch(ch -> ch == '0')) return "";
    return traceId;
  }

  static String newTraceId32() {
    return UUID.randomUUID().toString().replace("-", "");
  }

  static String safe(String s) {
    return Optional.ofNullable(s).orElse("");
  }
}
