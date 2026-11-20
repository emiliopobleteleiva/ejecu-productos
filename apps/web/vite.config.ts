import { defineConfig } from "vitest/config"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"

// App de desarrollo local de los microcomponentes de Productos.
// En dev las llamadas a /api las responde MSW (src/mocks), no la API real.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    // TODO: usar 51N0 cuando el documento de componentes nos asigne el número N.
    port: 5173,
    strictPort: true,
  },
  test: {
    environment: "jsdom",
    include: ["src/**/*.test.{ts,tsx}"],
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      reportsDirectory: "coverage",
      include: ["src/**/*.{ts,tsx}"],
      // Lo generado, el entrypoint y el arranque del worker de MSW no son código
      // propio que tenga sentido medir.
      exclude: [
        "src/api/**",
        "src/main.tsx",
        "src/mocks/navegador.ts",
        "src/**/*.test.{ts,tsx}",
      ],
    },
  },
})
