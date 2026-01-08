import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { rumVitePlugin } from '@arms/rum-vite-plugin'

export default defineConfig({
    plugins: [
        vue(),
        rumVitePlugin({
            pid: process.env.ARMS_RUM_PID || '',
            region: process.env.ARMS_RUM_REGION || 'us',
            version: process.env.APP_VERSION || '1.0.0',

            // 如果你将来要自动上传 SourceMap，再放开
            // accessKeyId: process.env.ALIYUN_ACCESS_KEY_ID,
            // accessKeySecret: process.env.ALIYUN_ACCESS_KEY_SECRET,
        }),
    ],

    build: {
        sourcemap: true,
    },

    server: {
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
        },
    },
})
