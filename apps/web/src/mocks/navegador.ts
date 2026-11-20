// Service worker de MSW para el navegador. Solo se carga en desarrollo
// (ver main.tsx). El archivo public/mockServiceWorker.js lo genera `msw init`.
import { setupWorker } from "msw/browser"
import { handlers } from "./handlers.ts"

export const trabajador = setupWorker(...handlers)
