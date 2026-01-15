import armsRum from "@arms/rum-browser";

let inited = false;

export function initRum() {
    if (inited) return armsRum;

    // ✅ build-time 注入：Vite 会在 npm run build 时写死到产物里
    const pid = import.meta.env.VITE_ARMS_RUM_PID;
    const endpoint = import.meta.env.VITE_ARMS_RUM_ENDPOINT;
    const appVersion = import.meta.env.VITE_APP_VERSION || "dev";

    console.log("[rum] pid=", pid);
    console.log("[rum] endpoint=", endpoint);
    console.log("[rum] appVersion=", appVersion);

    if (!pid || !endpoint) {
        console.warn(
            "[rum] missing pid/endpoint at build-time. " +
            "Did you pass --build-arg VITE_ARMS_RUM_PID / VITE_ARMS_RUM_ENDPOINT ?"
        );
        return armsRum;
    }

    armsRum.init({
        pid,
        endpoint,
        env: "prod",
        spaMode: "history",
        appVersion,

        collectors: {
            perf: true,
            webVitals: true,
            api: true,
            staticResource: true,
            jsError: true,
            consoleError: true,
            action: true,
        },

        tracing: {
            enable: true,
            sample: 100,
            allowedUrls: [{ match: "/api/", propagatorTypes: ["tracecontext"] }],
        },
    });

    inited = true;
    return armsRum;
}

export { armsRum };
