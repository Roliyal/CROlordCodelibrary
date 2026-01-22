package com.example.javasvc.config;

import com.example.javasvc.util.TraceIdUtil;
import jakarta.servlet.FilterChain;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.apache.logging.log4j.CloseableThreadContext;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.apache.logging.log4j.ThreadContext;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.net.InetAddress;

/**
 * Production-hardened HTTP access logging:
 * - Extract/Create TraceID
 * - Set unified MDC fields
 * - Log inbound start + done with costMs/status
 * - Ensure MDC is cleared to avoid leak
 * - Add X-Trace-Id response header for debugging
 */
@Component
public class HttpAccessLogFilter extends OncePerRequestFilter {

  private static final Logger log = LogManager.getLogger(HttpAccessLogFilter.class);

  private final ArmsLogProperties props;

  public HttpAccessLogFilter(ArmsLogProperties props) {
    this.props = props;
  }

  @Override
  protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain) {
    long start = System.nanoTime();
    String traceId = TraceIdUtil.extractOrCreateHttp(request);

    response.setHeader("X-Trace-Id", traceId);

    ThreadContext.put("service", "java-service");
    ThreadContext.put("env", props.getEnv());
    ThreadContext.put("version", props.getVersion());

    ThreadContext.put("traceId", traceId);
    ThreadContext.put("protocol", "http");
    ThreadContext.put("direction", "inbound");
    ThreadContext.put("method", request.getMethod());
    ThreadContext.put("path", request.getRequestURI());
    ThreadContext.put("peer", request.getRemoteAddr());

    try (var ctx = CloseableThreadContext.put("source", "HttpAccessLogFilter")
        .put("category", "http.inbound.start")) {

      log.info("http request start");
      filterChain.doFilter(request, response);

      long costMs = (System.nanoTime() - start) / 1_000_000;
      try (var ctx2 = CloseableThreadContext.put("category", "http.inbound.done")
          .put("costMs", String.valueOf(costMs))
          .put("status", String.valueOf(response.getStatus()))) {
        log.info("http request done");
      }

    } catch (Exception e) {
      long costMs = (System.nanoTime() - start) / 1_000_000;
      String stack = e.toString();
      try (var ctx = CloseableThreadContext.put("category", "http.inbound.error")
          .put("source", "HttpAccessLogFilter")
          .put("costMs", String.valueOf(costMs))
          .put("status", "500")
          .put("errorType", e.getClass().getName())
          .put("errorMessage", safe(e.getMessage()))
          .put("errorStack", stack)) {
        log.error("http request failed");
      }
      throw new RuntimeException(e);
    } finally {
      ThreadContext.clearAll();
    }
  }

  private static String safe(String s) {
    return s == null ? "" : s;
  }
}
