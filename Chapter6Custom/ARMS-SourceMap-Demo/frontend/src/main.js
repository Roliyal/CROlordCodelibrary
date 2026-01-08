// 解决部分库在浏览器中引用 global 的问题
globalThis.global ||= globalThis

import { createApp } from 'vue'
import App from './App.vue'
import armsRum from '@arms/rum-browser'

const pid = import.meta.env.VITE_ARMS_RUM_PID
const endpoint = import.meta.env.VITE_ARMS_RUM_ENDPOINT

// 启动时打印，确认 env 已注入
console.log('env pid=', pid)
console.log('env endpoint=', endpoint)

armsRum.init({
    pid,
    endpoint,

    env: 'prod',
    spaMode: 'history',
    appVersion: import.meta.env.VITE_APP_VERSION || 'dev',

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
        allowedUrls: [
            { match: '/api/', propagatorTypes: ['tracecontext'] },
        ],
    },
})

createApp(App).mount('#app')
