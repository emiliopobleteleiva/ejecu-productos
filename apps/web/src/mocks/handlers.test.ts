import { afterAll, beforeAll, describe, expect, it } from "vitest"
import { setupServer } from "msw/node"
import { handlers } from "./handlers.ts"
import { departamentos } from "./datos.ts"

// Los handlers se prueban contra el mismo contrato v1 que consume el front:
// SKU autogenerado por departamento, 422 si el departamento no existe, 404 si el
// producto no está y baja lógica en el DELETE.
const servidor = setupServer(...handlers)
const base = "http://localhost/api"

beforeAll(() => servidor.listen({ onUnhandledRequest: "error" }))
afterAll(() => servidor.close())

function cuerpoCreacion(departamentoId: string, nombre: string) {
  return JSON.stringify({
    departamento_id: departamentoId,
    nombre,
    categoria_ids: ["5b0c7e1a-1f2d-4c3b-8a9e-6d7f8a9b0c01"],
    costo_unitario: 1000,
    tipo_impuesto: "IVA",
    permite_venta: true,
  })
}

describe("mocks del contrato v1", () => {
  it("lista los departamentos ordenados por código", async () => {
    const respuesta = await fetch(`${base}/departamentos`)
    expect(respuesta.status).toBe(200)

    const cuerpo = (await respuesta.json()) as {
      departamentos: { codigo: string }[]
      total: number
    }
    expect(cuerpo.total).toBe(departamentos.length)

    const codigos = cuerpo.departamentos.map((d) => d.codigo)
    expect(codigos).toEqual([...codigos].sort())
  })

  it("filtra los productos por departamento", async () => {
    const primero = departamentos[0]
    expect(primero).toBeDefined()
    const respuesta = await fetch(`${base}/productos?departamento_id=${primero!.id}`)
    expect(respuesta.status).toBe(200)

    const cuerpo = (await respuesta.json()) as {
      productos: { departamento_id: string }[]
      total: number
    }
    expect(cuerpo.total).toBe(2)
    expect(cuerpo.productos.every((p) => p.departamento_id === primero!.id)).toBe(true)
  })

  it("responde 422 con problem+json si el departamento no existe", async () => {
    const respuesta = await fetch(`${base}/productos`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: cuerpoCreacion("00000000-0000-4000-8000-000000000000", "Fantasma"),
    })

    expect(respuesta.status).toBe(422)
    expect(respuesta.headers.get("content-type")).toContain("application/problem+json")
    const problema = (await respuesta.json()) as { detail: string }
    expect(problema.detail).toContain("departamento")
  })

  it("crea un producto con SKU autogenerado a partir del departamento", async () => {
    const coleccionables = departamentos[1]
    expect(coleccionables).toBeDefined()

    const respuesta = await fetch(`${base}/productos`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: cuerpoCreacion(coleccionables!.id, "Carta coleccionable"),
    })

    expect(respuesta.status).toBe(201)
    const creado = (await respuesta.json()) as { sku: string; departamento_id: string }
    // COL ya tenía COL-000001 en los datos iniciales.
    expect(creado.sku).toBe("COL-000002")
    expect(creado.departamento_id).toBe(coleccionables!.id)
  })

  it("actualiza solo los campos enviados y mantiene el SKU", async () => {
    const primero = departamentos[0]
    expect(primero).toBeDefined()

    const lista = await fetch(`${base}/productos?departamento_id=${primero!.id}`)
    const { productos } = (await lista.json()) as {
      productos: { id: string; sku: string }[]
    }
    const objetivo = productos[0]
    expect(objetivo).toBeDefined()

    const respuesta = await fetch(`${base}/productos/${objetivo!.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ costo_unitario: 19990 }),
    })

    expect(respuesta.status).toBe(200)
    const actualizado = (await respuesta.json()) as {
      sku: string
      costo_unitario: number
    }
    expect(actualizado.costo_unitario).toBe(19990)
    expect(actualizado.sku).toBe(objetivo!.sku)
  })

  it("responde 404 al actualizar un producto inexistente", async () => {
    const respuesta = await fetch(
      `${base}/productos/00000000-0000-4000-8000-000000000000`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nombre: "Nadie" }),
      },
    )

    expect(respuesta.status).toBe(404)
  })

  it("aplica baja lógica: después del DELETE el producto responde 404", async () => {
    const libro = departamentos[2]
    expect(libro).toBeDefined()

    const creacion = await fetch(`${base}/productos`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: cuerpoCreacion(libro!.id, "Tomo nuevo"),
    })
    const creado = (await creacion.json()) as { id: string }

    const borrado = await fetch(`${base}/productos/${creado.id}`, { method: "DELETE" })
    expect(borrado.status).toBe(204)

    const consulta = await fetch(`${base}/productos/${creado.id}`)
    expect(consulta.status).toBe(404)
  })
})
