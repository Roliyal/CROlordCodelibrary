import os
import time
from flask import Flask, request, jsonify
from dotenv import load_dotenv

from opentelemetry import trace, propagate
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter

from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
import requests


def must_env(key: str) -> str:
    v = os.getenv(key, "")
    if not v:
        raise RuntimeError(f"missing env {key}")
    return v


def init_otel(service_name: str):
    otlp_endpoint = must_env("OTEL_EXPORTER_OTLP_ENDPOINT")
    provider = TracerProvider(resource=Resource.create({"service.name": service_name}))

    exporter = OTLPSpanExporter(
        endpoint=otlp_endpoint,
        headers={},
        timeout=10,
    )
    provider.add_span_processor(BatchSpanProcessor(exporter))
    trace.set_tracer_provider(provider)


def cur_ids():
    span = trace.get_current_span()
    sc = span.get_span_context() if span else None
    if sc and sc.is_valid:
        return format(sc.trace_id, "032x"), format(sc.span_id, "016x")
    return "", ""


def log(prefix: str, extra: str = ""):
    tid, sid = cur_ids()
    msg = f"{prefix} trace_id={tid} span_id={sid}"
    if extra:
        msg += f" {extra}"
    print(msg, flush=True)


# ---- boot: load env from baked file first ----
dotenv_path = os.getenv("DOTENV_PATH", "/app/.env")
loaded = load_dotenv(dotenv_path, override=True)
print(f"[py] load_dotenv path={dotenv_path} loaded={loaded}", flush=True)

# 兼容本地开发：如果你在 python-svc 目录直接跑，也可继续读本地/上级 .env
load_dotenv(".env", override=True)
load_dotenv("../.env", override=True)

init_otel("python-svc")

app = Flask(__name__)
FlaskInstrumentor().instrument_app(app)
RequestsInstrumentor().instrument()

PY_PORT = int(os.getenv("PY_PORT", "8081"))
JAVA_URL = os.getenv("JAVA_URL", "")


@app.get("/healthz")
def healthz():
    log("[py] /healthz")
    return "ok", 200


@app.get("/py/work")
def py_work():
    tp_in = request.headers.get("traceparent", "")
    log("[py] /py/work", f"traceparent_in={tp_in}")

    downstream_java = None
    if JAVA_URL:
        try:
            headers = {}
            propagate.inject(headers)  # inject tracecontext
            r = requests.get(f"{JAVA_URL}/java/work", headers=headers, timeout=5)
            downstream_java = r.json()
        except Exception as e:
            downstream_java = {"error": str(e)}
    else:
        downstream_java = {"warn": "JAVA_URL is empty"}

    tid, sid = cur_ids()
    return jsonify({
        "message": "hello from python",
        "time": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
        "trace_id": tid,
        "span_id": sid,
        "traceparent_in": tp_in,
        "java": downstream_java,
    })


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=PY_PORT, debug=False)
