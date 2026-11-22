// Atajos a los esquemas del contrato. Los tipos salen de contrato.d.ts, que se
// genera con `pnpm web:tipos` desde contracts/openapi.yaml: no se escriben a mano.
import type { components } from "./contrato"

type Esquemas = components["schemas"]

export type Producto = Esquemas["Producto"]
export type Departamento = Esquemas["Departamento"]
export type CrearProducto = Esquemas["CrearProductoBody"]
export type ActualizarProducto = Esquemas["ActualizarProductoBody"]
export type ListaProductos = Esquemas["ListarProductosBody"]
export type ListaDepartamentos = Esquemas["ListarDepartamentosBody"]
export type Problema = Esquemas["ErrorModel"]
