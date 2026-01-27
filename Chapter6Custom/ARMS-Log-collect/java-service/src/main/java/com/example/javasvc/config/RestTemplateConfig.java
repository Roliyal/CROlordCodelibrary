package com.example.javasvc.config;

import org.apache.logging.log4j.ThreadContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.HttpRequest;
import org.springframework.http.client.*;

import java.io.IOException;
import java.time.Duration;

/**
 * Production: enforce timeouts + trace propagation.
 */
@Configuration
public class RestTemplateConfig {

  @Bean
  public org.springframework.web.client.RestTemplate restTemplate(ArmsLogProperties props) {
    var factory = new SimpleClientHttpRequestFactory();
    int timeoutMs = (int) Math.max(1, props.getHttpTimeoutMillis());
    factory.setConnectTimeout(timeoutMs);
    factory.setReadTimeout(timeoutMs);

    var rt = new org.springframework.web.client.RestTemplate(factory);
    rt.getInterceptors().add(new TracePropagateInterceptor());
    return rt;
  }

  static class TracePropagateInterceptor implements ClientHttpRequestInterceptor {
    @Override
    public ClientHttpResponse intercept(HttpRequest request, byte[] body, ClientHttpRequestExecution execution)
        throws IOException {

      String traceId = ThreadContext.get("traceId");
      if (traceId != null && !traceId.isBlank()) {
        request.getHeaders().set("X-Trace-Id", traceId);
        // Optional W3C traceparent (minimal demo): version 00 + dummy span
        request.getHeaders().set("traceparent", "00-" + traceId + "-0000000000000000-01");
      }
      return execution.execute(request, body);
    }
  }
}
