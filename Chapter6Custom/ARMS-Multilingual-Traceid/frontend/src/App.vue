<script setup>
import { ref } from "vue";
import { armsRum } from "./rum";

const result = ref("");

async function callApi() {
  result.value = "calling /api/hello ...";
  const res = await fetch("/api/hello");
  const data = await res.json();
  result.value = JSON.stringify(data, null, 2);
}

function throwJsError() {
  throw new Error("ARMS_SOURCEMAP_DEMO_FROM_APP_VUE");
}

function sendBizException() {
  armsRum.sendException({
    name: "ORDER_CREATE_FAILED",
    message: "库存不足，创建订单失败",
    properties: {
      orderId: "order_123",
      sku: "sku_001",
      reason: "out_of_stock",
    },
  });
  result.value = "sendException called";
}

function sendBizCustomEvent() {
  armsRum.sendCustom({
    type: "biz",
    name: "order_create_failed",
    group: "order",
    value: 1,
    properties: {
      orderId: "order_123",
      sku: "sku_001",
      amount: 99,
      currency: "CNY",
    },
  });
  result.value = "sendCustom called";
}
</script>

<template>
  <div style="padding: 24px; font-family: Arial">
    <h2>ARMS Multilingual TraceId – RUM + OTEL</h2>

    <button @click="callApi">Call /api/hello (Go → Py → Java → C++)</button>
    <br /><br />

    <button @click="throwJsError">1️ Throw JS Error（SourceMap）</button>
    <br /><br />

    <button @click="sendBizException">2️ sendException（业务异常）</button>
    <br /><br />

    <button @click="sendBizCustomEvent">3️ sendCustom（自定义事件）</button>

    <pre
        style="
        margin-top: 16px;
        background: #111;
        color: #0f0;
        padding: 12px;
        border-radius: 8px;
        white-space: pre-wrap;
      "
    >
{{ result }}
    </pre>
  </div>
</template>
