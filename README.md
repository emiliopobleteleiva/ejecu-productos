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
- Levanta con `docker compose up -d --build` (API + PostgreSQL).
- Hooks locales de formato, `go vet` y Gitleaks con `lefthook install`.
- Sin autenticación todavía; los permisos quedan anotados como `TODO`.
- Sin migraciones, sqlc ni conexión a PostgreSQL.

## Pipeline de CI

El workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml) corre en cada
PR y en cada push a `main`. Los tres jobs son checks requeridos en `main`.

| Job | Controles |
|---|---|
| `calidad` | `go mod verify`, vet y build, contrato al día, cobertura y SonarCloud con Quality Gate |
| `seguridad` | gitleaks, govulncheck, pnpm audit, hadolint y Trivy fs |
| `imagen` | build de la imagen, Trivy image y Dockle |

- Ejecuciones del pipeline:
  [GitHub Actions](https://github.com/emiliopobleteleiva/ejecu-productos/actions/workflows/ci.yml)
- Análisis de código:
  [SonarCloud](https://sonarcloud.io/dashboard?id=emiliopobleteleiva_ejecu-productos)

## Ramas y commits

`main` está protegida: todo entra por PR, con una aprobación y los checks
`calidad`, `seguridad` e `imagen` en verde.

| Prefijo | Uso |
|---|---|
| `feat/<tema>` | funcionalidad nueva |
| `fix/<tema>` | corrección de un defecto |
| `chore/<tema>` | tooling, pipeline, dependencias y configuración |
| `docs/<tema>` | solo documentación |

Los commits siguen Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`, `ci:`).
`chore/` y `docs/` se agregan a los prefijos del DAS (§9) para el trabajo que no
cambia el comportamiento del servicio.

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
docker-compose.yaml           # postgres + backend para desarrollo
lefthook.yml                  # hooks de pre-commit y pre-push
docs/modos-de-falla.md
docs/decisiones.md
docs/arquitectura.md
docs/entorno.md
```

## Requisitos

Las instrucciones completas para CachyOS/Arch y Ubuntu/Debian están en
[`docs/entorno.md`](docs/entorno.md).

Resumen:

- Go 1.26.0 + toolchain go1.26.11.
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

## Cómo levantar todo con Docker Compose

```bash
cp .env.example .env        # completar POSTGRES_PASSWORD
docker compose up -d --build
docker compose ps
```

| Servicio | URL |
|---|---|
| API | `http://localhost:5131/api/productos` |
| Documentación interactiva | `http://localhost:5131/api/docs` |
| PostgreSQL | `localhost:5433` |

`docker compose down` para bajar conservando los datos; `docker compose down -v`
borra también el volumen de PostgreSQL. Los puertos siguen la convención `51N1`
del DAS y se publican solo en `127.0.0.1`.

## Hooks locales

```bash
pnpm install
pnpm exec lefthook install
```

`lefthook` es una `devDependency` (2.1.10): `pnpm exec` lo resuelve desde
`node_modules`, así que no hace falta instalarlo aparte ni tenerlo en el PATH.
`.npmrc` tiene `ignore-scripts=true`, así que los hooks no se instalan solos. Sin
`pnpm exec lefthook install` no hay validación local: el `pre-commit` corre
gofmt, `go vet` y Gitleaks, y el `pre-push` corre los tests y verifica que el
contrato esté al día. Detalle y prueba de Gitleaks en
[`docs/entorno.md`](docs/entorno.md).

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

- Go: `1.26.0` + `toolchain go1.26.11`.
- Huma: `v2.34.1`.
- No se agregan dependencias fuera de la tabla del DAS.
