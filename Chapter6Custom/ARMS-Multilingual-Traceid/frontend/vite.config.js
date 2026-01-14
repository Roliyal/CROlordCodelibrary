import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";

export default defineConfig(({ mode }) => {
    const rootEnvDir = path.resolve(__dirname, "..");
    const env = loadEnv(mode, rootEnvDir, "VITE_");

    return {
        plugins: [vue()],
        envDir: rootEnvDir,           // ✅ dev 时从外部 .env 读
        build: { sourcemap: true },
        server: {
            proxy: {
                "/api": { target: "http://127.0.0.1:8080", changeOrigin: true },
            },
        },
    };
});
