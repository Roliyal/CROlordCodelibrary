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
  public org.springframework.web.client.RestTemplate restTemplate() {
    var factory = new SimpleClientHttpRequestFactory();
    factory.setConnectTimeout((int) Duration.ofSeconds(2).toMillis());
    factory.setReadTimeout((int) Duration.ofSeconds(3).toMillis());

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
