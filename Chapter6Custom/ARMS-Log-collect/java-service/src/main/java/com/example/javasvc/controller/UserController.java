package com.example.javasvc.controller;

import com.example.javasvc.grpc.GoBridgeGrpcClient;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.LinkedHashMap;
import java.util.Map;
import java.util.UUID;

@RestController
public class UserController {

    private final GoBridgeGrpcClient goGrpc;

    public UserController(GoBridgeGrpcClient goGrpc) {
        this.goGrpc = goGrpc;
    }

    @GetMapping("/api/user/get")
    public ResponseEntity<Map<String, Object>> getUser(
            @RequestParam(name = "n", defaultValue = "1") int n,
            @RequestHeader(name = "X-Trace-Id", required = false) String traceHeader
    ) {
        // IMPORTANT: do NOT call go-service HTTP endpoints from here.
        // Otherwise you'll create a loop:
        // go /api/payment/pay -> java /api/user/get -> go /api/payment/query -> java /api/order/create -> go /api/payment/pay ...
        String traceId = (traceHeader != null && !traceHeader.isBlank())
                ? traceHeader
                : UUID.randomUUID().toString().replace("-", "");

        Map<String, Object> resp = new LinkedHashMap<>();
        resp.put("traceId", traceId);
        resp.put("userId", "u-" + n);
        resp.put("ok", true);

        // optional: small, safe gRPC call to go-service (no cycle)
        try {
            // NOTE: keep response JSON-friendly (avoid putting protobuf objects directly)
            var r = goGrpc.queryPayment(traceId, "{\"q\":\"u-" + n + "\"}");
            resp.put("goGrpc", r.getResult());
        } catch (Exception e) {
            // still return 200, because this endpoint is a dependency of go-service /api/payment/pay
            resp.put("goGrpcError", e.getMessage());
        }

        return ResponseEntity.ok(resp);
    }
}
