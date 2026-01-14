import { armsRum } from "./rum";

let lastBackendTrace = { trace_id: "", span_id: "", time: "" };

export function getLastBackendTrace() {
    return { ...lastBackendTrace };
}

export function logWithTrace(message, extra = {}) {
    const ctx = getLastBackendTrace();
    const payload = { ...extra, backend_trace_id: ctx.trace_id, backend_span_id: ctx.span_id };
    console.log(`[fe] ${message} trace_id=${ctx.trace_id} span_id=${ctx.span_id}`, payload);
}

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
            backend_time: ctx.time,
        },
    });
}

export function sendBizException(name, message, properties = {}) {
    const ctx = getLastBackendTrace();
    armsRum.sendException({
        name,
        message,
        properties: {
            ...properties,
            backend_trace_id: ctx.trace_id,
            backend_span_id: ctx.span_id,
            backend_time: ctx.time,
        },
    });
}

export async function callHelloApi({ manualTraceparent = "" } = {}) {
    const headers = {};
    if (manualTraceparent) headers["traceparent"] = manualTraceparent;

    const res = await fetch("/api/hello", { method: "GET", headers });
    const data = await res.json();

    lastBackendTrace = {
        trace_id: data?.trace_id || "",
        span_id: data?.span_id || "",
        time: data?.time || "",
    };

    logWithTrace("call /api/hello done", { status: res.status });
    sendBizEvent("api_hello_ok", { status: res.status });

    return data;
}
