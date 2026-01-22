package com.example.javasvc.controller;

import org.apache.logging.log4j.CloseableThreadContext;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.http.*;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

/**
 * Production: ensure errors are logged as structured JSON and response is consistent.
 */
@RestControllerAdvice
public class GlobalExceptionHandler {
  private static final Logger log = LogManager.getLogger(GlobalExceptionHandler.class);

  @ExceptionHandler(Exception.class)
  public ResponseEntity<Map<String, Object>> handle(Exception e) {
    try (var ctx = CloseableThreadContext.put("source", "GlobalExceptionHandler")
        .put("category", "http.unhandled_exception")
        .put("errorType", e.getClass().getName())
        .put("errorMessage", safe(e.getMessage()))
        .put("errorStack", e.toString())) {
      log.error("unhandled exception");
    }

    return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
        .contentType(MediaType.APPLICATION_JSON)
        .body(Map.of("code", 500, "message", "internal error"));
  }

  private static String safe(String s) { return s == null ? "" : s; }
}
