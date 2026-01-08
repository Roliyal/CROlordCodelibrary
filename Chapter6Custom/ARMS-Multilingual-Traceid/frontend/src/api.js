import { armsRum } from "./rum";

// 当前页面“最近一次后端 trace 上下文”
let lastBackendTrace = {
    trace_id: "",
    span_id: "",
    time: ""
};

export function getLastBackendTrace() {
    return { ...lastBackendTrace };
}

// 统一日志：每条都带 trace_id/span_id（如果有）
export function logWithTrace(message, extra = {}) {
    const ctx = getLastBackendTrace();
    const payload = { ...extra, backend_trace_id: ctx.trace_id, backend_span_id: ctx.span_id };
    console.log(`[fe] ${message} trace_id=${ctx.trace_id} span_id=${ctx.span_id}`, payload);
}

// 统一发送自定义事件：把 trace_id/span_id 写进去，便于 ARMS 里筛选/检索
export function sendBizEvent(name, properties = {}) {
    const ctx = getLastBackendTrace();
    armsRum.sendCustom({
        type: "biz",
        name,
        group: "demo",
        value: 1,
        properties: {
            ...properties,
            backend_trace_id: ctx.trace_id,
            backend_span_id: ctx.span_id,
            backend_time: ctx.time
        }
    });
}

// 统一发送异常：同样把 trace_id/span_id 写进去
export function sendBizException(name, message, properties = {}) {
    const ctx = getLastBackendTrace();
    armsRum.sendException({
        name,
        message,
        properties: {
            ...properties,
            backend_trace_id: ctx.trace_id,
            backend_span_id: ctx.span_id,
            backend_time: ctx.time
        }
    });
}

// ✅ 关键：调用后端接口，更新 lastBackendTrace
export async function callHelloApi({ manualTraceparent = "" } = {}) {
    const headers = {};

    // （可选）手动注入 traceparent：一般不需要，RUM 已经会自动注入
    // 但当你想“固定 traceparent 做对照测试”时可以用。
    if (manualTraceparent) {
        headers["traceparent"] = manualTraceparent;
    }

    const res = await fetch("/api/hello", {
        method: "GET",
        headers
    });

    const data = await res.json();

    // 你的 Go 返回里应当有 trace_id/span_id
    lastBackendTrace = {
        trace_id: data?.trace_id || "",
        span_id: data?.span_id || "",
        time: data?.time || ""
    };

    logWithTrace("call /api/hello done", { status: res.status });
    sendBizEvent("api_hello_ok", { status: res.status });

    return data;
}
