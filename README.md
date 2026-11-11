# ejecu-productos

Grupo 3 pingeso

Microservicio de productos del curso DAS. Implementa el Hito 1: contrato
publicado y endpoints respondiendo con datos simulados. PostgreSQL, sqlc,
goose y la lógica real quedan para el Hito 2.

## Estado actual

- CRUD de productos con datos en memoria.
- SKU único, categorización múltiple con `categoria_ids []string`, costo
  `int64` en pesos y flag `permite_venta`.
- Contrato OpenAPI generado desde los structs de Huma.
- Sin autenticación todavía; los permisos quedan anotados como `TODO`.
- Sin migraciones, sqlc ni conexión a PostgreSQL.

## Estructura

```text
apps/api/
├── cmd/api/                  # arranque del servidor
├── cmd/codegen/              # genera contracts/openapi.yaml
├── internal/api/             # structs Huma, registro y handlers simulados
├── internal/service/         # TODO Hito 2: reglas de negocio
├── internal/store/           # TODO Hito 2: PostgreSQL
└── go.mod
contracts/openapi.yaml        # generado, no editar a mano
docs/modos-de-falla.md
docs/decisiones.md
docs/entorno.md
```

## Requisitos

Las instrucciones completas para CachyOS/Arch y Ubuntu/Debian están en
[`docs/entorno.md`](docs/entorno.md).

Resumen:

- Go 1.26.0 + toolchain go1.26.5.
- Node 22.22.0 y pnpm 10.13.1, solo para el flujo del DAS.
- Docker y Docker Compose para el Hito 2.
- `go install` deja los binarios en `$HOME/go/bin`; ese directorio debe estar
  en el `PATH`. En fish:

  ```fish
  fish_add_path "$HOME/go/bin"
  fish_add_path "$HOME/.local/bin"
  exec fish
  ```

- Dockle se usa como imagen: `goodwithtech/dockle:v0.4.15` (el tag lleva `v`).

## Endpoints

| Método | Ruta | Operación |
|---|---|---|
| `POST` | `/api/productos` | Crear producto |
| `GET` | `/api/productos` | Listar productos con filtros `categoria_id` y `permite_venta` |
| `GET` | `/api/productos/{id}` | Obtener producto por UUID |
| `PATCH` | `/api/productos/{id}` | Actualizar parcialmente |
| `DELETE` | `/api/productos/{id}` | Desactivar (baja lógica) |

En `API_ENV=dev` la documentación interactiva está en `/api/docs` y el
contrato en `/api/openapi.json`.

## Cómo levantar la API

Desde `apps/api`:

```bash
go mod tidy
go run ./cmd/api
```

Variables opcionales para desarrollo:

```bash
API_PORT=8080
API_ENV=dev
```

Probar:

```bash
curl -s http://localhost:8080/api/productos | jq
curl -s http://localhost:8080/api/docs -o /dev/null -w '%{http_code}\n'
```

## Cómo generar el contrato

Desde `apps/api`:

```bash
go run ./cmd/codegen
```

El archivo se escribe en `contracts/openapi.yaml`. No se edita a mano: se
regenera desde los structs.

## Decisiones

- El modelado de categorías múltiples está en
  [`docs/decisiones.md`](docs/decisiones.md).
- Los modos de falla y su verificación están en
  [`docs/modos-de-falla.md`](docs/modos-de-falla.md).
- El entorno de desarrollo está en
  [`docs/entorno.md`](docs/entorno.md).

## Versiones fijadas

- Go: `1.26.0` + `toolchain go1.26.5`.
- Huma: `v2.34.1`.
- No se agregan dependencias fuera de la tabla del DAS.
