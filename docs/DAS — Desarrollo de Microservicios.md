# DAS — Desarrollo de Microservicios

## 1\. Introducción

Todo microservicio del curso, de Civil o de Ejecución, se construye sobre la misma plantilla técnica base. Este documento fija lo que no se discute: stack, estructura de repositorio, contenedores, contrato API y los controles de seguridad que corren desde el commit hasta la imagen.

**Alcance.** Lineamientos técnicos transversales. Qué servicios existen, qué hace cada uno y de cuáles depende está en el documento de componentes de cada sección, no aquí.

**Audiencia.** Todos los grupos de desarrollo y los jefes de proyecto que revisan cumplimiento.

**Regla general.** Lo que este documento marca como obligatorio lo verifica el pipeline. Un PR que no pasa el pipeline no se integra, sin excepciones manuales.

**Referencias**

- Documento de componentes de cada sección (Civil y Ejecución): servicios, responsabilidades, dependencias y mocks  
- Hitos de entrega del curso, comunes a ambas secciones

## 2\. Arquitectura y contrato API

Cada microservicio tiene su propio repositorio, su propia base de datos, su propio pipeline y un contrato OpenAPI publicado. Ningún servicio accede a los datos de otro si no es por su API.

### Reglas de arquitectura

- **Dueño único por dato.** Cada dato lo escribe un solo servicio. Los demás lo piden por API.  
- **Dependencias en un solo sentido.** No hay ciclos entre servicios. Si dos servicios se necesitan mutuamente, el límite entre ellos está mal trazado.  
- **API-first.** Un servicio publica su contrato antes que su lógica, para que quien depende de él trabaje en paralelo con mocks.  
- **Clean Architecture.** Handler → Service → Store dentro de cada servicio (sección 3).  
- **Microcomponentes.** Cada servicio entrega su frontend como microcomponentes, que después compone cada grupo cliente (sección 3).

### Comunicación entre servicios

- **REST síncrono con JSON**, siempre contra el contrato publicado del otro servicio. Mensajería asíncrona solo si el documento de componentes de la sección la define.  
- Un servicio llama a otro por su URL interna (`<SERVICIO>_URL`), reenviando el JWT del usuario y el `X-Request-Id`.  
- **Una operación que cambia datos de otro servicio se confirma solo si ese servicio la confirma.** La llamada lleva `Idempotency-Key`, para que un reintento no la aplique dos veces. La reversa es una operación inversa con su propia clave.  
- Timeouts explícitos en todo cliente HTTP (5 s por defecto). Un servicio caído devuelve un error claro, no cuelga la request.

### Contrato API-first

El contrato se publica antes que la lógica. Con Huma el contrato nace de los structs Go, así que API-first significa: primero structs y handlers que devuelven datos simulados, después la lógica real.

1. Definir structs de request y response en `internal/api/` con tags `json`, validaciones y ejemplos.  
2. `pnpm codegen` genera `contracts/openapi.yaml` (OpenAPI 3.1) y los tipos TypeScript. Ambos se commitean.  
3. El servicio consumidor copia el `openapi.yaml` del proveedor a `contracts/externos/<servicio>.yaml` y genera desde ahí su cliente (openapi-typescript en el front, oapi-codegen en Go).  
4. Mientras el proveedor no tenga lógica, el consumidor trabaja contra `prism mock contracts/externos/<servicio>.yaml` o contra handlers MSW.  
5. Un cambio incompatible (quitar un campo, cambiar un tipo, volver obligatorio un parámetro) lo bloquea oasdiff en CI. Si es necesario, se publica `/api/v2/...` y se avisa a los consumidores antes de retirar `v1`.

En `API_ENV=dev` cada servicio sirve su documentación interactiva en `/api/docs`.

## 3\. Stack tecnológico y estructura de repositorio

Todos los servicios parten de la plantilla base: mismo stack, mismas versiones, misma estructura. Cambiar una pieza del stack requiere aprobación del jefe de proyecto y queda registrado aquí.

El frontend es React y la base de datos es PostgreSQL en todas las secciones.

Las versiones de abajo son exactas, sin rangos. Subir una versión es un PR propio, que pasa por el pipeline completo y actualiza esta tabla.

### Backend (Go)

| Componente | Versión | Dónde se fija | Uso |
| :---- | :---- | :---- | :---- |
| Go (toolchain) | 1.26.5 | `go.mod`: `go 1.26.0` \+ `toolchain go1.26.5` y la imagen de build | Todos |
| `github.com/danielgtaylor/huma/v2` | v2.34.1 | `go.mod` | HTTP \+ OpenAPI 3.1 |
| `github.com/jackc/pgx/v5` | v5.10.0 | `go.mod` | Driver PostgreSQL |
| `github.com/sqlc-dev/sqlc` | v1.29.0 | `Makefile` (`go run ...@v1.29.0`) | Generación de queries |
| `github.com/pressly/goose/v3` | v3.24.1 | `go.mod`, `Makefile`, `Dockerfile.migrate` | Migraciones |
| `github.com/riverqueue/river` (+ `riverpgxv5`, `rivertype`) | v0.40.0 | `go.mod` | Jobs, solo si el servicio los usa |
| `github.com/golang-jwt/jwt/v5` | v5.3.1 | `go.mod` | Emisión (servicio de identidad) y validación de JWT |
| `github.com/MicahParks/keyfunc/v3` | v3.9.1 | `go.mod` | Lectura y caché del JWKS |
| `golang.org/x/crypto` | v0.54.0 | `go.mod` | Hash de contraseñas, solo servicio de identidad |
| `github.com/oapi-codegen/oapi-codegen/v2` | v2.8.2 | `Makefile` (`go run ...@v2.8.2`) | Cliente Go de servicios externos |
| `github.com/testcontainers/testcontainers-go/modules/postgres` | v0.43.0 | `go.mod` | PostgreSQL real en pruebas |
| `github.com/wneessen/go-mail` | v0.8.1 | `go.mod` | SMTP, solo si el servicio envía correo |

### Frontend

| Componente | Versión | Dónde se fija |
| :---- | :---- | :---- |
| Node.js | 22.22.0 | `.nvmrc` y `engines` de `package.json` y la imagen de build |
| pnpm | 10.13.1 | `packageManager` de `package.json` |
| react / react-dom | 19.2.8 | catalog de `pnpm-workspace.yaml` |
| typescript | 6.0.3 | catalog |
| vite | 8.1.5 | catalog |
| @vitejs/plugin-react | 6.0.4 | `apps/web/package.json` |
| tailwindcss / @tailwindcss/vite | 4.3.3 | `apps/web/package.json` |
| @base-ui/react (shadcn, estilo `base-nova`) | 1.6.0 | `apps/web/package.json`, `components.json` |
| react-router | 8.3.0 | `apps/web/package.json` |
| @tanstack/react-query | 5.101.4 | `apps/web/package.json` |
| react-hook-form | 7.83.0 | `apps/web/package.json` |
| openapi-fetch | 0.15.2 | `apps/web/package.json` |
| openapi-typescript | 7.13.0 | `apps/web/package.json` |
| msw | 2.15.0 | `apps/web/package.json` |
| vitest | 4.1.10 | `apps/web/package.json` |
| @playwright/test | 1.62.0 | `apps/web/package.json` |
| oxlint | 1.75.0 | `package.json` raíz |
| prettier / prettier-plugin-tailwindcss | 3.9.6 / 0.8.1 | `package.json` raíz |
| lefthook | 2.1.10 | `package.json` raíz |

### Imágenes base y servicios de Compose

| Imagen | Tag | Digest |
| :---- | :---- | :---- |
| golang (build) | `1.26.5-alpine3.23` | `sha256:622e56dbc11a8cfe87cafa2331e9a201877271cbff918af53d3be315f3da88cc` |
| node (build web) | `22.22.0-alpine3.23` | `sha256:e4bf2a82ad0a4037d28035ae71529873c069b13eb0455466ae0bc13363826e34` |
| distroless (runtime api y web) | `static-debian13:nonroot` | `sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6` |
| alpine (runtime migrate) | `3.23.7` | Se fija en el Hito 0 |
| postgres | `17.11` | Se fija en el Hito 0 |
| axllent/mailpit (solo perfil dev) | `v1.31.3` | Se fija en el Hito 0 |

Para obtener un digest: `docker buildx imagetools inspect postgres:17.11 --format '{{json .Manifest.Digest}}'`. Se usa el digest del índice multi-arquitectura, no el de una plataforma.

### Reglas de fijación

- `.npmrc` con `save-exact=true`: `pnpm add` escribe `19.2.8`, no `^19.2.8`. El catalog también va sin `^`.  
- `pnpm install --frozen-lockfile` en CI y en Docker, y `go mod verify` en CI. Si el lockfile no coincide, falla.  
- Ninguna herramienta se invoca con `@latest` ni con un tag flotante como `postgres:17` o `alpine:3.23`.  
- Las GitHub Actions se fijan por SHA de commit, con la versión en un comentario (sección 6).  
- Dependabot o Renovate abre los PR de actualización. Nadie sube versiones a mano dentro de un PR de funcionalidad.

### Estructura de cada repo

```
<servicio>/
├── apps/
│   ├── api/
│   │   ├── cmd/api/            # main: arranque, config, señales
│   │   ├── internal/api/       # handlers Huma + structs del contrato
│   │   ├── internal/service/   # reglas de negocio, sin HTTP ni SQL
│   │   ├── internal/store/     # acceso a PostgreSQL
│   │   ├── internal/repository/# generado por sqlc, no editar
│   │   ├── db/migrations/      # goose
│   │   ├── db/queries/         # SQL para sqlc
│   │   ├── e2e/                # e2e + testdata/*.golden.md
│   │   ├── Dockerfile
│   │   └── Dockerfile.migrate
│   └── web/
│       └── src/features/<feature>/
├── contracts/openapi.yaml      # generado y commiteado
├── docs/modos-de-falla.md
├── docker-compose.yaml
├── .env.example
├── lefthook.yml
└── .github/workflows/ci.yml
```

### Capas del backend

Handler → Service → Store. El handler solo traduce HTTP, el service no conoce HTTP ni SQL y el store no tiene reglas de negocio. Un handler que llama al store directamente se rechaza en revisión.

### Frontend: microcomponentes

Cada servicio entrega su frontend como microcomponentes: componentes React autocontenidos, cada uno con su propio cliente de API tipado. La base común queda formada por las APIs y estos microcomponentes, y cada grupo cliente los compone después en su propio producto.

- Un microcomponente no importa a otro ni depende de un layout, una ruta o un estado global.  
- Recibe por props la URL base de su API y lo que necesite del contexto (usuario, permisos).  
- Estructura: `src/componentes/<kebab-case>/` con `<nombre>.tsx` (vista, sin lógica de negocio), `<nombre>.api.ts` (único lugar con `fetch`) y `use-<nombre>.ts` (estado).  
- Se publica como paquete npm versionado (`@<org>/<servicio>-ui`) y cada versión declara qué versión del contrato API consume.  
- El repo incluye una app de desarrollo local que monta los microcomponentes con mocks MSW. Esa app no se publica.

## 4\. Contenedorización

Cada servicio publica dos imágenes, `api` y `migrate`, y su paquete de microcomponentes. Las imágenes se construyen con build multi-stage, imágenes base fijadas por digest y runtime sin shell ni compilador. Todos los servicios usan los mismos Dockerfiles y solo cambia el nombre del binario.

### Reglas obligatorias (las verifican Hadolint, Dockle y Trivy)

| Regla | Cómo se cumple |
| :---- | :---- |
| Base fijada por versión y digest | `golang:1.26.5-alpine3.23@sha256:…`, nunca `latest` |
| Runtime mínimo | `gcr.io/distroless/static-debian13:nonroot` para api y web |
| Usuario no root | `USER nonroot:nonroot`, archivos con `--chown` |
| Binario reproducible | `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, `-ldflags="-s -w -X main.version=$VERSION"` |
| Caché de dependencias | `RUN --mount=type=cache` para módulos Go y store de pnpm |
| Multi-arquitectura | `--platform=$BUILDPLATFORM` \+ `TARGETOS`/`TARGETARCH` |
| Sin secretos en la imagen | `.dockerignore` excluye `.env*`, `.git`, `node_modules`. Ningún secreto en `ARG` ni en `VITE_*` |
| Un puerto | Todo escucha en `8080` dentro del contenedor |
| Healthcheck | API: `GET /healthz` (vivo) y `GET /readyz` (con PostgreSQL). Web: `GET /healthz` |

### API (`apps/api/Dockerfile`)

```dockerfile
# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM golang:1.26.5-alpine3.23@sha256:622e56dbc11a8cfe87cafa2331e9a201877271cbff918af53d3be315f3da88cc AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY apps/api/ ./
ARG TARGETOS TARGETARCH VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -buildvcs=false -trimpath \
    -ldflags="-s -w -X main.version=$VERSION" \
    -o /out/servicio ./cmd/api

FROM gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6
WORKDIR /app
COPY --from=build --chown=nonroot:nonroot /out/servicio /app/servicio
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/servicio"]
```

&nbsp;

### Migraciones (`apps/api/Dockerfile.migrate`)

Goose v3.24.1 se instala con `go install github.com/pressly/goose/v3/cmd/goose@v3.24.1` para que use su propio `go.mod`. La imagen final es `alpine:3.23.7` (digest fijado en el Hito 0\) con `goose`, `river-migrate` (si el servicio usa River) y `db/migrations/`. No contiene Go ni necesita internet. Comando: `goose -dir /app/db/migrations postgres "$DATABASE_URL" up`.

### Web (`apps/web/Dockerfile`)

Sirve la app de desarrollo local de cada repo y es la plantilla para la app que componga los microcomponentes. Tres etapas: build de Vite con `pnpm install --frozen-lockfile --ignore-scripts`, compilación del `spa-server` en Go y runtime distroless. El `spa-server` sirve `/srv`, responde `/healthz`, aplica `X-Content-Type-Options: nosniff` y `Referrer-Policy`, cachea `/assets/` como inmutable y reenvía `/api` al backend vía `BACKEND_URL`. Así el navegador ve un solo origen y la cookie de sesión funciona sin CORS.

### Compose de desarrollo

`docker-compose.yaml` en la raíz con `name:` fijo y estos servicios:

1. `postgres` con healthcheck `pg_isready`  
2. `migrate`, que espera `service_healthy` de postgres y termina  
3. `backend`, que espera `service_completed_successfully` de migrate, con puerto publicado solo en `127.0.0.1`  
4. `frontend`, con `BACKEND_URL=http://backend:8080`  
5. Perfil `dev`: `seed` (datos de demo) y `mailpit` (SMTP falso en `:8025`). Fuera de desarrollo no se define `COMPOSE_PROFILES`

Cada servicio recibe un número N en el documento de componentes de su sección y usa un rango de puertos locales propio, para que un desarrollador pueda levantar varios a la vez: `51N0` (web) y `51N1` (api). Ejemplo: el servicio 2 usa 5120 y 5121\.

## 5\. Datos, migraciones, configuración y secretos

Cada servicio tiene su propia base PostgreSQL y nadie más se conecta a ella. Si un servicio necesita datos de otro, los pide a su API. Un `JOIN` contra la base de otro servicio se rechaza.

### Migraciones

- Toda modificación de esquema es una migración goose nueva en `db/migrations/`. Una migración ya desplegada no se edita.  
- Las queries viven en `db/queries/*.sql` y `make generate` regenera `internal/repository/` con sqlc. Ese directorio no se edita a mano.  
- La API nunca migra al arrancar. Corre primero el contenedor `migrate` y después la API.  
- Los montos se guardan como enteros en pesos (`bigint`), nunca como `float`. Las fechas, como `timestamptz`.  
- Los identificadores expuestos en la API son UUID.

### Configuración

Toda la configuración entra por variables de entorno, se valida al arrancar y el proceso termina si falta una obligatoria. `.env.example` lista todas las variables con valores `replace-with-*` y `.env` nunca se commitea.

| Variable | Uso | Obligatoria |
| :---- | :---- | :---- |
| `API_PORT` | Puerto interno, siempre 8080 | Sí |
| `API_ENV` | `dev`, `test` o `prod`. En `prod` no se publican `/api/docs` ni el spec | Sí |
| `DATABASE_URL` | Conexión a la base propia. Dentro de Compose apunta al nombre del servicio, no a `localhost` | Sí |
| `AUTH_JWKS_URL` | Llaves públicas del servicio de identidad para validar tokens | Sí, salvo el servicio de identidad |
| `<SERVICIO>_URL` | URL de cada servicio del que se depende (ej. `IDENTIDAD_URL`) | Según dependencias |
| `FRONTEND_URL` | Origen permitido y base de enlaces | Sí |
| `TZ` | `America/Santiago` | Sí |

### Secretos

- Ningún secreto en el repositorio, en la imagen, en `ARG` de Docker ni en variables `VITE_*` (todo `VITE_*` termina en el bundle público).  
- En CI los secretos van en GitHub Actions Secrets. En local van en un `.env` fuera del control de versiones.  
- Gitleaks bloquea el commit si detecta un secreto (sección 6). Si uno se filtra igual, se rota de inmediato: borrarlo del historial no basta.  
- La sesión viaja en cookie `HttpOnly`, `Secure`, `SameSite=Lax`. `localStorage` y `sessionStorage` están prohibidos en todo el frontend.

## 6\. Seguridad y shift-left

Cada control corre lo más temprano posible: lo que se puede detectar en el editor no espera al commit, y lo que se puede detectar en el commit no espera al pipeline. El pipeline repite todo lo anterior, porque los hooks locales se pueden saltar.

### Controles por etapa

| Etapa | Herramienta | Qué detecta | Bloquea |
| :---- | :---- | :---- | :---- |
| Editor | `.editorconfig`, gofmt, Prettier, oxlint | Formato y reglas de código | No (aviso) |
| pre-commit | Prettier, oxlint `--deny-warnings`, gofmt, `go vet` | Formato, `fetch` fuera de `*.api.ts`, `localStorage`, imports entre features | Sí |
| pre-commit | Gitleaks `protect --staged` | Secretos en lo que se va a commitear | Sí |
| pre-push | `pnpm typecheck`, `pnpm test`, `go test ./...` | Tipos y pruebas | Sí |
| CI: código | Lint, typecheck, build, e2e con golden | Regresiones funcionales | Sí |
| CI: contrato | `git diff --exit-code` tras `pnpm codegen` y oasdiff contra `main` | Contrato desactualizado o cambio incompatible | Sí |
| CI: dependencias | govulncheck, `pnpm audit --audit-level high` | Vulnerabilidades alcanzables en Go y npm | Sí |
| CI: repo | `trivy fs --scanners vuln,misconfig,secret` | CVE en lockfiles, Dockerfile y Compose mal configurados, secretos | Sí, HIGH/CRITICAL |
| CI: Dockerfile | Hadolint | Malas prácticas en Dockerfiles | Sí, nivel `error` |
| CI: calidad | SonarQube \+ Quality Gate | Bugs, vulnerabilidades, code smells, cobertura en código nuevo | Sí |
| CI: imagen | `trivy image`, Dockle | CVE del runtime, root, secretos en capas | Sí, HIGH/CRITICAL |
| CI: publicación | SBOM CycloneDX (Trivy), firma con Cosign | Trazabilidad de lo publicado | No, se adjunta a la release |

### Hooks locales (`lefthook.yml`)

```yaml
pre-commit:
  parallel: true
  commands:
    format:
      glob: "*.{js,ts,jsx,tsx,json,css,md,yml,yaml}"
      run: pnpm exec prettier --write --ignore-unknown {staged_files}
      stage_fixed: true
    lint:
      run: pnpm exec oxlint --deny-warnings
    go-fmt:
      glob: "*.go"
      run: cd apps/api && test -z "$(gofmt -l .)" && go vet ./...
    secretos:
      run: gitleaks protect --staged --redact --no-banner

pre-push:
  parallel: true
  commands:
    typecheck:
      run: pnpm typecheck
    test:
      run: pnpm test
    test-api:
      run: pnpm test:api
```

&nbsp;

`.npmrc` tiene `ignore-scripts=true`, así que después de `pnpm install` hay que correr `pnpm exec lefthook install` a mano. Sin eso no hay hooks.

### Pipeline de CI (`.github/workflows/ci.yml`)

Corre en cada PR y en cada push a `main`. Los jobs `calidad`, `seguridad` e `imagen` son checks requeridos en la protección de rama.

```yaml
name: ci
on:
  pull_request:
  push:
    branches: [main]

permissions:
  contents: read

env:
  TRIVY_VERSION: v0.76.1
  GOVULNCHECK_VERSION: v1.9.0
  OASDIFF_VERSION: v1.35.0
  DOCKLE_VERSION: v0.4.15

jobs:
  calidad:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with: { fetch-depth: 0 }
      - uses: pnpm/action-setup@d9184bf108216479bc5a137cc391f4d7b14c870b # v6.1.0
      - uses: actions/setup-node@949feb2413d6458794dcd2491c4babbbce0c15c1 # v7.1.0
        with: { node-version-file: .nvmrc, cache: pnpm }
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with: { go-version-file: apps/api/go.mod }
      - run: pnpm install --frozen-lockfile
      - run: cd apps/api && go mod verify
      - run: pnpm verify
      - run: pnpm codegen && git diff --exit-code
      - name: contrato compatible con main
        run: |
          git show origin/main:contracts/openapi.yaml > /tmp/base.yaml
          go run github.com/oasdiff/oasdiff@${OASDIFF_VERSION} breaking /tmp/base.yaml contracts/openapi.yaml --fail-on ERR
      - run: pnpm test:e2e:api   # binario con -cover, deja cobertura en GOCOVERDIR
      - uses: SonarSource/sonarqube-scan-action@4d1dc0c91ee400dcf8a6dddf2f705ac36542e304 # v8.4.0
        env:
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
          SONAR_HOST_URL: ${{ secrets.SONAR_HOST_URL }}
      - uses: SonarSource/sonarqube-quality-gate-action@7a5fffe8e523c40e0c740b6bc2712ab503e52efa # v1.2.1
        env:
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}

  seguridad:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with: { fetch-depth: 0 }
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with: { go-version-file: apps/api/go.mod }
      - uses: gitleaks/gitleaks-action@e0c47f4f8be36e29cdc102c57e68cb5cbf0e8d1e # v3.0.0
        env:
          GITLEAKS_VERSION: 8.30.1
      - run: cd apps/api && go run golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION} ./...
      - uses: hadolint/hadolint-action@06be81baf89a55ffd0e24b8f04a4185738dd3387 # v3.5.0 (hadolint v2.15.1)
        with: { recursive: true, failure-threshold: error }
      - uses: aquasecurity/trivy-action@a9c7b0f06e461e9d4b4d1711f154ee024b8d7ab8 # v0.36.0
        with:
          version: ${{ env.TRIVY_VERSION }}
          scan-type: fs
          scanners: vuln,misconfig,secret
          severity: HIGH,CRITICAL
          ignore-unfixed: true
          exit-code: "1"

  imagen:
    needs: [calidad, seguridad]
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - run: docker build --pull -f apps/api/Dockerfile --build-arg VERSION=${GITHUB_SHA::7} -t api:ci .
      - run: docker build --pull -f apps/web/Dockerfile -t web:ci .
      - uses: aquasecurity/trivy-action@a9c7b0f06e461e9d4b4d1711f154ee024b8d7ab8 # v0.36.0
        with: { version: "${{ env.TRIVY_VERSION }}", image-ref: api:ci, severity: "HIGH,CRITICAL", ignore-unfixed: true, exit-code: "1" }
      - uses: aquasecurity/trivy-action@a9c7b0f06e461e9d4b4d1711f154ee024b8d7ab8 # v0.36.0
        with: { version: "${{ env.TRIVY_VERSION }}", image-ref: web:ci, severity: "HIGH,CRITICAL", ignore-unfixed: true, exit-code: "1" }
      - run: docker run --rm -v /var/run/docker.sock:/var/run/docker.sock goodwithtech/dockle:${DOCKLE_VERSION} --exit-code 1 --exit-level warn api:ci
```

&nbsp;

Cada action va fijada por SHA de commit con su versión en comentario. Un tag se puede mover, un SHA no.

### Versiones de herramientas de seguridad y CI

| Herramienta | Versión | Fijada en |
| :---- | :---- | :---- |
| Trivy | v0.76.1 | `TRIVY_VERSION` del workflow, y local con la misma versión |
| aquasecurity/trivy-action | v0.36.0 (`a9c7b0f0…`) | workflow |
| Gitleaks | v8.30.1 | `GITLEAKS_VERSION` del workflow, y local para el hook |
| gitleaks/gitleaks-action | v3.0.0 (`e0c47f4f…`) | workflow |
| Hadolint (vía hadolint-action v3.5.0) | v2.15.1 | workflow |
| Dockle | v0.4.15 | `DOCKLE_VERSION` del workflow |
| govulncheck (`golang.org/x/vuln`) | v1.9.0 | `GOVULNCHECK_VERSION` del workflow |
| oasdiff | v1.35.0 | `OASDIFF_VERSION` del workflow |
| Prism (`@stoplight/prism-cli`) | 5.16.0 | `devDependencies` de `apps/web` |
| Cosign | v3.1.4 | job de publicación |
| SonarSource/sonarqube-scan-action | v8.4.0 (`4d1dc0c9…`) | workflow |
| SonarSource/sonarqube-quality-gate-action | v1.2.1 (`7a5fffe8…`) | workflow |
| actions/checkout | v7.0.1 (`3d3c42e5…`) | workflow |
| actions/setup-go | v7.0.0 (`b7ad1dad…`) | workflow |
| actions/setup-node | v7.1.0 (`949feb24…`) | workflow |
| pnpm/action-setup | v6.1.0 (`d9184bf1…`) | workflow |
| Runner | `ubuntu-24.04` | workflow (no `ubuntu-latest`) |

### Umbrales

- **Trivy:** 0 hallazgos HIGH o CRITICAL con fix disponible, en repo e imágenes. Una excepción va en `.trivyignore.yaml` con el CVE, el motivo y una fecha de revisión, aprobada por el jefe de proyecto.  
- **SonarQube (Quality Gate sobre código nuevo):** 0 bugs y 0 vulnerabilidades nuevas, 100% de security hotspots revisados, duplicación bajo 3% y cobertura de al menos 60%. La cobertura sale de los e2e: la API se compila con `go build -cover` y el reporte se convierte con `go tool covdata textfmt`.  
- **Gitleaks:** cero hallazgos. No hay excepciones por archivo.

### Controles de aplicación

- Autenticación: El servicio de identidad emite un JWT firmado (EdDSA o RS256) en cookie `HttpOnly`. Los demás servicios lo validan con `AUTH_JWKS_URL`. Nadie comparte la llave privada.  
- Autorización: cada endpoint declara su permiso (ej. `<recurso>:administrar`, `<recurso>:operar`) y existe una matriz de acceso probada por rol en `acceso_test.go`.  
- Toda entrada se valida en el contrato (tipos, `minLength`, `maximum`, `pattern`). Huma rechaza lo que no cumple antes de llegar al service.  
- Operaciones que mueven cupos, inventario o dinero aceptan `Idempotency-Key`, para que un reintento no descuente dos veces.  
- Los errores siguen RFC 9457 (`application/problem+json`) y nunca exponen trazas, SQL ni rutas internas.

## 7\. Escáneres obligatorios: SonarQube y Trivy

Cada repo corre los dos escáneres en cada PR y ninguno reemplaza al otro. SonarQube mira el código propio y Trivy mira todo lo que el código trae de afuera.

### Por qué dos escáneres

| Escáner | Analiza | Detecta | No cubre en esta configuración |
| :---- | :---- | :---- | :---- |
| SonarQube | Código propio de `apps/api` y `apps/web`, pruebas y cobertura | Bugs, vulnerabilidades, code smells, duplicación, security hotspots y cobertura del código nuevo | CVE de dependencias, imagen base, Dockerfile y secretos |
| Trivy | Dependencias y lockfiles, Dockerfiles, Compose, imágenes y SBOM | CVE con corrección disponible, malas configuraciones y secretos embebidos | Lógica del código, calidad y cobertura |

Un CVE en una dependencia no lo detecta el análisis de código. Un error de lógica o una consulta SQL mal armada no lo detecta el escaneo de dependencias e imágenes.

### Qué debe solicitar cada grupo

La solicitud se hace al jefe de proyecto antes de cerrar el Hito 0\. Sin proyecto, token y jobs configurados no hay Quality Gate ni evidencia para el hito.

| Solicitud | Cuándo | Evidencia |
| :---- | :---- | :---- |
| Proyecto SonarQube con clave estable (ej. `<servicio>`) | Antes del primer PR | URL del proyecto |
| Token de análisis y `SONAR_HOST_URL` | Antes del primer PR | Secrets `SONAR_TOKEN` y `SONAR_HOST_URL` cargados en el repo |
| `sonar-project.properties` en la raíz | Antes del primer PR | Archivo commiteado |
| Jobs `calidad`, `seguridad` e `imagen` activos | Hito 0 | Pipeline en verde y checks requeridos en `main` |
| Excepción de Trivy, si aplica | Antes del merge | `.trivyignore.yaml` con CVE, motivo y fecha de revisión, aprobado |

Para la solicitud basta indicar servicio, repo, clave del proyecto SonarQube, URL de la instancia y responsable del grupo. El token es un secreto: no se comparte por chat ni por correo y se carga directo como secret del repositorio.

### SonarQube: cómo se habilita y cómo se ejecuta

1. Crear o solicitar el proyecto con una clave estable, por ejemplo `<servicio>`. La clave no cambia entre ramas ni entre hitos.  
2. Generar un token de análisis y guardarlo como `SONAR_TOKEN`, y la URL de la instancia como `SONAR_HOST_URL`. Ambos van en GitHub Actions Secrets, nunca en el repo.  
3. Definir "New Code" antes del primer análisis. Sin esa definición, el Quality Gate no mide el código nuevo.  
4. Commitear un `sonar-project.properties` en la raíz con `sonar.projectKey`, `sonar.sources` y las rutas de cobertura. Si las pruebas viven en carpetas propias, agregar `sonar.tests` sin solapar con `sources`.  
5. Generar la cobertura que SonarQube va a leer. En Go sale de los e2e: compilar la API con `go build -cover`, ejecutar las pruebas con `GOCOVERDIR` y convertir el resultado con `go tool covdata textfmt -i "$GOCOVERDIR" -o coverage.out`. En el front, vitest emite el `lcov`. Sin reporte de cobertura, el Quality Gate no puede cumplir el 60% de código nuevo.  
6. Agregar el scan y el Quality Gate al job `calidad` (sección 6). La action del gate falla si el gate está rojo, y esa falla es la que bloquea el merge.  
7. Antes de abrir el PR, correr el escáner local con las mismas propiedades (comando abajo). El token va en una variable de entorno local, no en el archivo.  
8. Si el gate falla, se corrige la causa. No se baja el umbral, no se excluye un archivo completo y no se marca un hotspot como seguro sin leerlo.

```properties
sonar.projectKey=<servicio>
sonar.sources=apps/api,apps/web/src
sonar.go.coverage.reportPaths=coverage.out
sonar.javascript.lcov.reportPaths=apps/web/coverage/lcov.info
sonar.exclusions=apps/api/internal/repository/**,apps/web/src/components/ui/**
```

&nbsp;

```sh
sonar-scanner -Dsonar.host.url="$SONAR_HOST_URL" -Dsonar.token="$SONAR_TOKEN"
```

### Trivy: cómo se habilita y cómo se ejecuta

Trivy no requiere alta ni token. Se solicita que el repo ejecute los jobs `seguridad` e `imagen` y que la protección de rama los exija. El escaneo se hace siempre con la versión fijada v0.76.1, nunca con `latest`.

1. Job `seguridad`: `trivy fs` sobre el repo con `--scanners vuln,misconfig,secret`, `--severity HIGH,CRITICAL`, `--ignore-unfixed` y `--exit-code 1`. Detecta CVE en `go.sum` y `pnpm-lock.yaml`, malas configuraciones de Dockerfile y Compose, y secretos.  
2. Job `imagen`: construir `api:ci` y `web:ci` y correr `trivy image` sobre cada una. Detecta CVE de la imagen base y de las capas, y secretos embebidos.  
3. Publicación: generar el SBOM CycloneDX y firmar la imagen. No bloquea el merge, pero deja trazabilidad de lo publicado.  
4. Antes del PR, repetir el escaneo local con la misma versión (comandos abajo).  
5. Si hay un hallazgo con corrección disponible, se actualiza la dependencia o la imagen base y se reconstruye. No se baja la severidad ni se ignora el CVE.  
6. Si no hay corrección, `--ignore-unfixed` ya lo deja fuera del bloqueo. Si se quiere dejar registro, se documenta, pero no se agrega una excepción innecesaria.  
7. Una excepción real va en `.trivyignore.yaml` con el CVE, el motivo y una fecha de revisión, aprobada por el jefe de proyecto. Una excepción sin fecha no se acepta.

```sh
trivy fs --scanners vuln,misconfig,secret --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 .
docker build -t api:local -f apps/api/Dockerfile .
trivy image --scanners vuln,secret --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 api:local
docker build -t web:local -f apps/web/Dockerfile .
trivy image --scanners vuln,secret --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 web:local
```

### Cómo leer y cerrar los hallazgos

| Resultado | Acción | No hacer |
| :---- | :---- | :---- |
| SonarQube con gate rojo | Corregir el bug o la vulnerabilidad, agregar la prueba que falta o bajar la duplicación | Bajar el umbral o excluir el archivo completo |
| Security hotspot | Revisarlo y justificar la decisión | Marcarlo como seguro sin leerlo |
| Cobertura menor a 60% en código nuevo | Agregar un e2e que ejercite el cambio | Excluir código nuevo del análisis |
| Trivy HIGH/CRITICAL con fix | Actualizar la dependencia o la imagen base | Ignorar el CVE o bajar la severidad |
| Trivy HIGH/CRITICAL sin fix | No bloquea por `--ignore-unfixed`, dejar registro si aplica | Agregar una excepción innecesaria |
| Secreto detectado | Rotarlo de inmediato y sacarlo del historial | Borrar solo el último commit |
| Falso positivo | Excepción con CVE, motivo y fecha | Excepción permanente o sin aprobación |

### Evidencia para el Hito 0

- [ ] Proyecto SonarQube creado, con `SONAR_TOKEN` y `SONAR_HOST_URL` cargados como secrets  
- [ ] `sonar-project.properties` commiteado con las rutas de cobertura  
- [ ] Pipeline con los jobs `calidad`, `seguridad` e `imagen` en verde  
- [ ] Enlace al pipeline y al proyecto SonarQube en el README o en la entrega del grupo

## 8\. Observabilidad, ambientes y publicación

Cada servicio se puede operar sin leer su código: expone salud, escribe logs estructurados y propaga un id de request. La imagen y la configuración son las mismas en todos los ambientes.

### Observabilidad mínima

- **Logs:** `log/slog` en JSON a stdout, con `servicio`, `version`, `request_id`, `metodo`, `ruta`, `estado` y `duracion_ms`. Nunca se loguean contraseñas, tokens, cookies ni el cuerpo completo de requests.  
- **Correlación:** el gateway asigna `X-Request-Id` si no viene. Cada servicio lo propaga en las llamadas que hace a otros servicios.  
- **Salud:** `GET /healthz` responde si el proceso vive y `GET /readyz` responde 200 solo si PostgreSQL contesta. Ninguno requiere autenticación.  
- **Auditoría:** las operaciones que cambian datos de negocio registran quién, qué y cuándo en una tabla propia del servicio.

### Ambientes

| Ambiente | Dónde | Datos | `API_ENV` | Docs `/api/docs` |
| :---- | :---- | :---- | :---- | :---- |
| Local | `docker compose` con perfil `dev` | Seed de demo \+ Mailpit | `dev` | Sí |
| CI | Testcontainers en GitHub Actions | Base nueva por test | `test` | Sí |

### Publicación

- Cada versión publica en GitHub Container Registry las imágenes `api` y `migrate` que pasaron el pipeline, etiquetadas con el SHA corto del commit (`ghcr.io/<org>/<servicio>-api:<sha>`). Nada se reconstruye fuera del pipeline.  
- Los microcomponentes se publican como paquete npm versionado en GitHub Packages (`@<org>/<servicio>-ui`).  
- Cada release incluye el `openapi.yaml` de esa versión, el SBOM y las notas de cambios. Es lo que consumen los grupos cliente.  
- El orden de arranque es siempre `migrate` y después `api`. Si `migrate` falla, la API no parte.  
- El ambiente productivo de cada producto lo define cada grupo cliente y queda fuera de este documento.

### Hitos

| Hito | Fecha | Entregable verificable |
| :---- | :---- | :---- |
| 0\. Entorno y seguridad | Semana del 2 de noviembre | Repo creado desde la plantilla, pipeline en verde con Trivy y SonarQube |
| 1\. Implementación inicial | 13 de noviembre | Contrato publicado y validado, endpoints respondiendo con mocks. Revisión con nota de la arquitectura |
| 2\. Refinamiento y calidad | 4 de diciembre | Lógica real sin mocks, persistencia en PostgreSQL, umbrales de Trivy y SonarQube superados |
| Ensamblaje LPS | 21 de diciembre | APIs y microcomponentes de la base común listos para que cada grupo cliente los componga. Presentación de la derivación por cliente |
| Cierre | 11 de enero | Validación con clientes y notas finales |

## 9\. Convenciones de desarrollo y checklist de inicio

### Convenciones

- **Idioma:** identificadores y comentarios en español, en Go y en TypeScript. Endpoints en español y plural: `/api/usuarios`, `/api/usuarios/{id}/roles`.  
- **Ramas:** `main` protegida. Trabajo en `feat/<tema>`, `fix/<tema>`. Merge solo por PR con al menos una aprobación de otra persona del grupo y checks en verde.  
- **Commits:** Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`).  
- **Formato:** Prettier sin punto y coma, ancho 90, y gofmt en Go. Lo aplica el hook.  
- **Dependencias JS:** versiones en el catalog de `pnpm-workspace.yaml`, nunca sueltas en `package.json`.  
- **Componentes UI:** `src/components/ui/` es generado por la CLI de shadcn y no se edita. Las variantes van en un wrapper en `src/components/`.  
- **Fechas:** la lógica con fechas se prueba también con `TZ=UTC` y `TZ=Pacific/Auckland`.

### Pruebas

E2E es el mecanismo principal y, en lo posible, el único. Reglas:

1. Antes de programar algo que se prueba aislado, se escribe `docs/modos-de-falla.md`: cada falla posible, qué la impide y cómo se verifica.  
2. El e2e levanta el binario real de la API contra PostgreSQL real (Testcontainers o `TEST_DATABASE_URL`).  
3. El escenario es de dificultad media o alta: concurrencia, rechazos esperados, estados intermedios, operaciones encadenadas. Ejemplo: dos operaciones simultáneas sobre el último cupo disponible, y una debe ser rechazada.  
4. El e2e termina en un reporte determinista (sin ids ni fechas absolutas) comparado con un golden commiteado en `e2e/testdata/`.  
5. Antes de aceptar el e2e se rompe a propósito la protección (lock, validación, constraint) y se confirma que falla.

### Checklist de inicio por grupo (Hito 0\)

- [ ] Crear `<servicio>` desde la plantilla y renombrar módulo Go, binario e imágenes  
- [ ] Proteger `main`: PR obligatorio, 1 aprobación, checks `calidad`, `seguridad` e `imagen` requeridos  
- [ ] `pnpm install`, `pnpm exec lefthook install`, instalar Gitleaks local y verificar que un commit con un secreto falso se bloquea  
- [ ] Copiar `.env.example` a `.env`, completar `replace-with-*` y levantar `docker compose up -d --build`  
- [ ] Crear el proyecto en SonarQube y cargar `SONAR_TOKEN` y `SONAR_HOST_URL` como secrets del repo  
- [ ] Primer pipeline en verde en `main`  
- [ ] Escribir `docs/modos-de-falla.md` del servicio  
- [ ] Borrador del contrato: structs en `internal/api/`, `pnpm codegen`, `contracts/openapi.yaml` publicado para los grupos que dependen de este servicio  
- [ ] Endpoints respondiendo con datos simulados (para el Hito 1\)