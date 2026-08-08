import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, ".", "VITE_");
  return {
    base: env.VITE_BASE || "/",
    plugins: [vue()],
    server: {
      port: 5174,
      proxy: {
        "/api": {
          target: "http://localhost:8080",
          changeOrigin: true
        }
      }
    }
  };
});
