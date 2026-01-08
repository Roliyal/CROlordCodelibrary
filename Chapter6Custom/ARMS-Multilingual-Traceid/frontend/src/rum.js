import armsRum from "@arms/rum-browser";

function getEnv(key, def = "") {
    return (window.__ENV__ && window.__ENV__[key]) || def;
}

export function initRum() {
    const pid = getEnv("VITE_ARMS_RUM_PID");
    const endpoint = getEnv("VITE_ARMS_RUM_ENDPOINT");
    const appVersion = getEnv("VITE_APP_VERSION", "dev");

    console.log("[rum] pid=", pid);
    console.log("[rum] endpoint=", endpoint);

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
            action: true
        },
        tracing: {
            enable: true,
            sample: 100,
            allowedUrls: [
                { match: "/api/", propagatorTypes: ["tracecontext"] }
            ]
        }
    });

    return armsRum;
}

export { armsRum };
