import { afterEach, describe, expect, it } from "vitest"
import { cleanup, render, screen } from "@testing-library/react"
import { App } from "./app.tsx"

afterEach(cleanup)

describe("App", () => {
  it("muestra el título de la app de desarrollo", () => {
    render(<App />)
    const titulo = screen.getByRole("heading", { level: 1 })
    expect(titulo.textContent).toContain("Productos")
  })

  it("explica que los microcomponentes se montan más adelante", () => {
    render(<App />)
    expect(screen.getByText(/microcomponentes/i)).toBeDefined()
  })
})
