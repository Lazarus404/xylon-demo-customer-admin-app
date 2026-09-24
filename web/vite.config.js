import { defineConfig } from "vite";

export default defineConfig({
  root: ".",
  base: "/",
  build: {
    outDir: "dist",
    emptyOutDir: true
  },
  esbuild: {
    jsx: "automatic",
    jsxImportSource: "strike-fw"
  },
  server: {
    port: 5174,
    proxy: {
      "/v1": {
        target: "http://127.0.0.1:8092"
      }
    }
  }
});
