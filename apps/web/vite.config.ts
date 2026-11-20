import { defineConfig } from "vite"
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
})
