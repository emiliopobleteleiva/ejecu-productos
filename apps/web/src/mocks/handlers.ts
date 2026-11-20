// Handlers MSW que imitan el contrato v1 de /api/productos y /api/departamentos.
// Reproducen las reglas que el front tiene que manejar: SKU autogenerado por
// departamento, 422 si el departamento no existe, 404 si el producto no está,
// y baja lógica en el DELETE.
import { http, HttpResponse } from "msw"
import type {
  ActualizarProducto,
  CrearProducto,
  ListaDepartamentos,
  ListaProductos,
  Problema,
  Producto,
} from "../api/tipos.ts"
import { departamentos, productosIniciales } from "./datos.ts"

// Copia de trabajo: los cambios viven en memoria y se pierden al recargar la página.
const productos: Producto[] = structuredClone(productosIniciales)

// Errores con forma RFC 9457 (application/problem+json), como responde Huma.
function problema(status: number, title: string, detail: string) {
  const cuerpo: Problema = { type: "about:blank", status, title, detail }
  return HttpResponse.json(cuerpo, {
    status,
    headers: { "Content-Type": "application/problem+json" },
  })
}

// Siguiente SKU del departamento: <CODIGO>-<6 dígitos>.
function siguienteSku(codigo: string) {
  const usados = productos.filter((p) => p.sku.startsWith(`${codigo}-`)).length
  return `${codigo}-${String(usados + 1).padStart(6, "0")}`
}

// Posición del producto activo con ese id, o -1 si no existe o está dado de baja.
function indiceActivo(id: string | readonly string[] | undefined) {
  return productos.findIndex((p) => p.id === id && p.activo)
}

export const handlers = [
  http.get("*/api/departamentos", () => {
    const cuerpo: ListaDepartamentos = {
      departamentos: [...departamentos].sort((a, b) => a.codigo.localeCompare(b.codigo)),
      total: departamentos.length,
    }
    return HttpResponse.json(cuerpo)
  }),

  http.get("*/api/productos", ({ request }) => {
    const filtros = new URL(request.url).searchParams
    const categoria = filtros.get("categoria_id")
    const departamento = filtros.get("departamento_id")
    const permiteVenta = filtros.get("permite_venta")

    const resultado = productos.filter(
      (p) =>
        p.activo &&
        (!categoria || p.categoria_ids.includes(categoria)) &&
        (!departamento || p.departamento_id === departamento) &&
        (permiteVenta === null || String(p.permite_venta) === permiteVenta),
    )
    const cuerpo: ListaProductos = { productos: resultado, total: resultado.length }
    return HttpResponse.json(cuerpo)
  }),

  http.post("*/api/productos", async ({ request }) => {
    const datos = (await request.json()) as CrearProducto
    const departamento = departamentos.find((d) => d.id === datos.departamento_id)
    if (!departamento) {
      return problema(422, "Unprocessable Entity", "El departamento no existe.")
    }
    const ahora = new Date().toISOString()
    const nuevo: Producto = {
      id: crypto.randomUUID(),
      sku: siguienteSku(departamento.codigo),
      departamento_id: datos.departamento_id,
      nombre: datos.nombre,
      descripcion: datos.descripcion,
      categoria_ids: datos.categoria_ids,
      costo_unitario: datos.costo_unitario,
      tipo_impuesto: datos.tipo_impuesto,
      permite_venta: datos.permite_venta,
      activo: true,
      creado_en: ahora,
      actualizado_en: ahora,
    }
    productos.push(nuevo)
    return HttpResponse.json(nuevo, { status: 201 })
  }),

  http.get("*/api/productos/:id", ({ params }) => {
    const i = indiceActivo(params.id)
    if (i < 0) return problema(404, "Not Found", "El producto no existe.")
    return HttpResponse.json(productos[i])
  }),

  http.patch("*/api/productos/:id", async ({ params, request }) => {
    const i = indiceActivo(params.id)
    const actual = productos[i]
    if (!actual) return problema(404, "Not Found", "El producto no existe.")
    const cambios = (await request.json()) as ActualizarProducto
    // El SKU y el departamento son inmutables: el contrato no los acepta aquí.
    const actualizado: Producto = {
      ...actual,
      ...cambios,
      actualizado_en: new Date().toISOString(),
    }
    productos[i] = actualizado
    return HttpResponse.json(actualizado)
  }),

  http.delete("*/api/productos/:id", ({ params }) => {
    const i = indiceActivo(params.id)
    const actual = productos[i]
    if (!actual) return problema(404, "Not Found", "El producto no existe.")
    // Baja lógica: el producto queda guardado pero deja de aparecer.
    productos[i] = { ...actual, activo: false }
    return new HttpResponse(null, { status: 204 })
  }),
]
