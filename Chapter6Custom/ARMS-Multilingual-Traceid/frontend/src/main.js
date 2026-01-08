globalThis.global ||= globalThis

import { createApp } from 'vue'
import App from './App.vue'
import ArmsRum from '@arms/rum-browser'

/**
 * ✅ 你根目录 .env 里已有：
 * VITE_ARMS_RUM_PID=...
 * VITE_ARMS_RUM_ENDPOINT=...
 * VITE_APP_VERSION=...
 *
 * Vite 只会把 VITE_ 前缀注入到前端环境。
 */
const pid = import.meta.env.VITE_ARMS_RUM_PID
const endpoint = import.meta.env.VITE_ARMS_RUM_ENDPOINT
const appVersion = import.meta.env.VITE_APP_VERSION || 'dev'

console.log('[rum] pid=', pid)
console.log('[rum] endpoint=', endpoint)

ArmsRum.init({
    pid,
    endpoint,
    env: 'prod',
    spaMode: 'history',
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

    /**
     * ✅ 关键：把 W3C tracecontext 注入到你的后端请求里
     * 让 go-gateway 能收到 traceparent，从而整条链路 trace_id 对齐。
     *
     * - allowedUrls：只对匹配的请求注入
     * - propagatorTypes：tracecontext = W3C traceparent
     */
    tracing: {
        enable: true,
        sample: 100,
        allowedUrls: [
            { match: '/api/', propagatorTypes: ['tracecontext'] }
        ]
    }
})

createApp(App).mount('#app')
