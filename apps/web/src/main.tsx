import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { App } from "./app.tsx"
import "./index.css"

// En desarrollo se enciende MSW antes de montar la app, para que la primera
// llamada a /api ya la responda el mock. El import es dinámico: en el build de
// producción esta rama se elimina y MSW no entra al bundle.
async function iniciarMocks() {
  if (!import.meta.env.DEV) return
  const { trabajador } = await import("./mocks/navegador.ts")
  await trabajador.start({ onUnhandledRequest: "bypass" })
}

const raiz = document.getElementById("raiz")
if (!raiz) throw new Error("Falta el elemento #raiz en index.html")

await iniciarMocks()

createRoot(raiz).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
