// Datos simulados para la app de desarrollo. Son de PyD Bazar para que la demo
// se entienda, pero el formato es el genérico del contrato v1:
// SKU = <código del departamento>-<6 dígitos>.
import type { Departamento, Producto } from "../api/tipos.ts"

export const departamentos: Departamento[] = [
  { id: "0d1f6a52-6b7e-4f3a-9a51-3c2f1e0a7b01", codigo: "JDM", nombre: "Juegos de mesa" },
  { id: "0d1f6a52-6b7e-4f3a-9a51-3c2f1e0a7b02", codigo: "COL", nombre: "Coleccionables" },
  { id: "0d1f6a52-6b7e-4f3a-9a51-3c2f1e0a7b03", codigo: "LIB", nombre: "Libros" },
]

export const categorias = {
  estrategia: "5b0c7e1a-1f2d-4c3b-8a9e-6d7f8a9b0c01",
  familiar: "5b0c7e1a-1f2d-4c3b-8a9e-6d7f8a9b0c02",
  figuras: "5b0c7e1a-1f2d-4c3b-8a9e-6d7f8a9b0c03",
  manga: "5b0c7e1a-1f2d-4c3b-8a9e-6d7f8a9b0c04",
} as const

const ahora = "2026-11-19T12:00:00Z"

// Estado inicial. Los handlers trabajan sobre una copia (ver handlers.ts).
export const productosIniciales: Producto[] = [
  {
    id: "9a7b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c01",
    sku: "JDM-000001",
    departamento_id: "0d1f6a52-6b7e-4f3a-9a51-3c2f1e0a7b01",
    nombre: "Catan",
    descripcion: "Juego de estrategia y comercio, 3 a 4 jugadores.",
    categoria_ids: [categorias.estrategia, categorias.familiar],
    costo_unitario: 32990,
    tipo_impuesto: "IVA",
    permite_venta: true,
    activo: true,
    creado_en: ahora,
    actualizado_en: ahora,
  },
  {
    id: "9a7b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c02",
    sku: "JDM-000002",
    departamento_id: "0d1f6a52-6b7e-4f3a-9a51-3c2f1e0a7b01",
    nombre: "Dixit",
    categoria_ids: [categorias.familiar],
    costo_unitario: 24990,
    tipo_impuesto: "IVA",
    permite_venta: true,
    activo: true,
    creado_en: ahora,
    actualizado_en: ahora,
  },
  {
    id: "9a7b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c03",
    sku: "COL-000001",
    departamento_id: "0d1f6a52-6b7e-4f3a-9a51-3c2f1e0a7b02",
    nombre: "Figura Funko edición limitada",
    categoria_ids: [categorias.figuras],
    costo_unitario: 15990,
    tipo_impuesto: "IVA",
    permite_venta: false,
    activo: true,
    creado_en: ahora,
    actualizado_en: ahora,
  },
  {
    id: "9a7b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c04",
    sku: "LIB-000001",
    departamento_id: "0d1f6a52-6b7e-4f3a-9a51-3c2f1e0a7b03",
    nombre: "Manga tomo 1",
    categoria_ids: [categorias.manga],
    costo_unitario: 8990,
    tipo_impuesto: "EXENTO",
    permite_venta: true,
    activo: true,
    creado_en: ahora,
    actualizado_en: ahora,
  },
]
