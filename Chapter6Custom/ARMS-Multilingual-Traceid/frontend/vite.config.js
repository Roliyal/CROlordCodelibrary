import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";

export default defineConfig(({ mode }) => {
    // 让 Vite 从“仓库根目录”读取 .env / .env.local / .env.[mode] / .env.[mode].local
    const rootEnvDir = path.resolve(__dirname, ".."); // ARMS-Multilingual-Traceid/

    // 只加载 VITE_ 前缀变量，保证能被 import.meta.env 访问
    const env = loadEnv(mode, rootEnvDir, "VITE_");

    return {
        plugins: [vue()],
        envDir: rootEnvDir, //  关键：外部 .env 的目录
        build: { sourcemap: true },

        // 可选：如果你想在 config 里用到 env，也可以 define 到客户端（不需要也行）
        define: {
            "import.meta.env.VITE_ARMS_RUM_PID": JSON.stringify(env.VITE_ARMS_RUM_PID),
            "import.meta.env.VITE_ARMS_RUM_ENDPOINT": JSON.stringify(env.VITE_ARMS_RUM_ENDPOINT),
            "import.meta.env.VITE_APP_VERSION": JSON.stringify(env.VITE_APP_VERSION)
        },

        server: {
            proxy: {
                "/api": {
                    target: "http://127.0.0.1:8080",
                    changeOrigin: true
                }
            }
        }
    };
});
