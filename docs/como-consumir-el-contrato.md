# Cómo consumir el contrato de productos

Guía para los grupos y servicios que dependen de `productos`. Sigue el modelo
API-first del DAS (§2): el consumidor no lee nuestro código, genera su cliente
desde `contracts/openapi.yaml`.

## Qué se publica

| Artefacto | Dónde | Nota |
|---|---|---|
| Contrato OpenAPI 3.1 | `contracts/openapi.yaml` en este repo | generado desde los structs y commiteado. **Es la fuente para generar clientes** |
| Spec en runtime | `GET /api/openapi.json` y `GET /api/openapi.yaml` | solo con `API_ENV=dev`; en `prod` no se publica |
| Documentación interactiva | `GET /api/docs` | solo con `API_ENV=dev` |
| API | `http://localhost:5131` en el compose de desarrollo | contenedor `backend` |

**No generes clientes contra un ambiente `prod`**: ahí el spec no se publica a
propósito. Usa siempre el `openapi.yaml` del repositorio o de una release.

## Checklist rápido

1. Copiar `contracts/openapi.yaml` a `contracts/externos/productos.yaml`.
2. Generar el cliente con la herramienta de tu lenguaje.
3. Desarrollar contra un mock (Prism o MSW) mientras no necesites datos reales.
4. Apuntar a la API con una variable de entorno, reenviando el JWT del usuario y
   `X-Request-Id`.

## Paso 1 — copiar el contrato

```bash
mkdir -p contracts/externos
cp /ruta/a/ejecu-productos/contracts/openapi.yaml contracts/externos/productos.yaml
```

Si no tienes el repo a mano, pídelo al grupo 3 o tómalo de la release
correspondiente. El archivo va versionado en tu repo: así el diff te muestra
cuándo cambió el contrato.

## Paso 2 — generar el cliente

### Go (servicios)

```bash
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.2 \
  -package clientesproductos -generate types,client \
  contracts/externos/productos.yaml > internal/clientes/productos.gen.go
```

Uso típico:

```go
cli, err := clientesproductos.NewClientWithResponses(cfg.ProductosURL)
if err != nil {
    return err
}

resp, err := cli.ObtenerProductoWithResponse(ctx, productoID)
switch {
case resp.JSON200 != nil:
    if !resp.JSON200.PermiteVenta {
        return ErrProductoNoVendible
    }
case resp.ApplicationproblemJSONDefault != nil:
    return fmt.Errorf("productos: %s", resp.ApplicationproblemJSONDefault.Detail)
}
```

Los nombres salen del contrato: `operationId: obtener-producto` genera
`ObtenerProductoWithResponse`, y `costo_unitario` genera `CostoUnitario`. Si
nosotros cambiamos un nombre o un tipo, tu build falla antes del merge.

### Frontend (TypeScript)

```bash
pnpm add -D openapi-typescript@7.13.0
pnpm add openapi-fetch@0.15.2
pnpm exec openapi-typescript contracts/externos/productos.yaml -o src/api/productos.d.ts
```

```ts
import createClient from "openapi-fetch"
import type { paths } from "./api/productos.d.ts"

const productos = createClient<paths>({ baseUrl: import.meta.env.VITE_PRODUCTOS_URL })

const { data, error } = await productos.GET("/api/productos", {
  params: { query: { permite_venta: true } },
})
```

## Paso 3 — desarrollar sin la API

Mientras no necesites datos reales, levanta un mock desde el mismo YAML:

```bash
npx @stoplight/prism-cli@5.16.0 mock contracts/externos/productos.yaml
```

Responde según los `example` del contrato. Si prefieres mocks en código, usa
`msw` con handlers equivalentes (es lo que hace `apps/web` de este repo).

## Paso 4 — apuntar a la API real

- La URL va por variable de entorno (`PRODUCTOS_URL`), nunca hardcodeada.
- Reenvía el **JWT del usuario** y el **`X-Request-Id`** en cada llamada (DAS §2).
- Define timeouts y trata los códigos de error (ver abajo).

## Qué garantiza el contrato y qué no

| Eje | Qué garantiza | Quién lo garantiza |
|---|---|---|
| Forma | Campos, tipos y nombres | El generador: si no compila, no cumple |
| Comportamiento | Códigos `404`/`422`, errores `problem+json`, baja lógica | Nuestros tests y [`modos-de-falla.md`](./modos-de-falla.md) |

El cliente generado no te avisa si nosotros devolvemos `200` donde correspondía
`404`. Eso lo cubren los tests de ambos lados.

## Contrato v1: lo que debes saber

Si venías del contrato anterior, estos cambios rompen clientes viejos:

| Cambio | Detalle |
|---|---|
| `sku` ya no se envía | Lo genera el servicio (`BEB-000001`) y es inmutable |
| `departamento_id` obligatorio | Catálogo en `GET /api/departamentos` |
| `costo` → `costo_unitario` | `int64` en pesos chilenos, IVA incluido |
| `tipo_impuesto` obligatorio | `enum`: `IVA` o `EXENTO` |
| Filtro nuevo | `departamento_id` en el listado de productos |

`sku` y `departamento_id` son inmutables: enviarlos en `POST` o `PATCH` devuelve
`422 unexpected property`.

## Versionado y cambios incompatibles

- Cada PR de este repo corre `oasdiff breaking` contra `main`: un cambio
  incompatible (quitar un campo, cambiar un tipo, volver obligatorio un parámetro
  opcional) se bloquea en CI.
- Si de verdad hay que romper el contrato, se publica `/api/v2/...` y se avisa a
  los consumidores antes de retirar `v1`.
- Cada release incluye el `openapi.yaml` de esa versión.

## Errores

Todos los errores usan RFC 9457 (`application/problem+json`) con el esquema
`ErrorModel`:

```json
{
  "title": "Unprocessable Entity",
  "status": 422,
  "detail": "El departamento indicado no existe.",
  "errors": [{ "location": "body.departamento_id", "message": "..." }]
}
```

No inventes un formato propio ni parsees el `detail` para tomar decisiones:
usa el `status` y, cuando exista, `errors[].location`.

## Contacto

Grupo 3 (servicio de productos). Cambios en el contrato se avisan en el PR o en
el canal del proyecto antes de mergear.
