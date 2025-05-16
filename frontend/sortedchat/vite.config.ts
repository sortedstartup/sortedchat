import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig(({ command }) => ({

  plugins: [tailwindcss(), reactRouter(), tsconfigPaths()],

  server: {
    proxy: command === 'serve' ? {
      '/hack': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/hack/, '')
      }
    } : undefined
  }

}));
