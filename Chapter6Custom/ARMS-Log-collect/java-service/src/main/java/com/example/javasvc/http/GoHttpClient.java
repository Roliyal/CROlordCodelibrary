package com.example.javasvc.http;

import com.example.javasvc.config.ArmsLogProperties;
import org.apache.logging.log4j.CloseableThreadContext;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

/**
 * Production: outbound HTTP client wrapper
 * - logs start/done/error
 * - records status + costMs
 * - propagates TraceID via RestTemplate interceptor
 */
@Component
public class GoHttpClient {
  private static final Logger log = LogManager.getLogger(GoHttpClient.class);

  private final RestTemplate rt;
  private final ArmsLogProperties props;

  public GoHttpClient(RestTemplate rt, ArmsLogProperties props) {
    this.rt = rt;
    this.props = props;
  }

  public String post(String path, String category) {
    String url = props.getGoHttpBaseUrl() + path;
    long start = System.nanoTime();

    try (var ctx = CloseableThreadContext.put("source", "GoHttpClient")
        .put("category", category)
        .put("protocol", "http")
        .put("direction", "outbound")
        .put("method", "POST")
        .put("path", path)
        .put("remoteService", "go-service")) {

      log.info("remote http call start");
      ResponseEntity<String> resp = rt.postForEntity(url, null, String.class);

      long costMs = (System.nanoTime() - start) / 1_000_000;
      try (var ctx2 = CloseableThreadContext.put("costMs", String.valueOf(costMs))
          .put("status", String.valueOf(resp.getStatusCode().value()))) {
        log.info("remote http call done");
      }
      return resp.getBody();

    } catch (Exception e) {
      long costMs = (System.nanoTime() - start) / 1_000_000;
      try (var ctx = CloseableThreadContext.put("source", "GoHttpClient")
          .put("category", "remote.http.error")
          .put("protocol", "http")
          .put("direction", "outbound")
          .put("method", "POST")
          .put("path", path)
          .put("remoteService", "go-service")
          .put("costMs", String.valueOf(costMs))
          .put("status", "ERR")
          .put("errorType", e.getClass().getName())
          .put("errorMessage", safe(e.getMessage()))
          .put("errorStack", e.toString())) {
        log.error("remote http call failed");
      }
      throw e;
    }
  }

  private static String safe(String s) { return s == null ? "" : s; }
}
