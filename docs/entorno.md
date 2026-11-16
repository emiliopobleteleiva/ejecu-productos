# Entorno de desarrollo

Requisitos para levantar el servicio y ejecutar las herramientas del DAS.
El repo se probó en CachyOS/Arch; los comandos para Ubuntu/Debian se incluyen
para el resto del grupo.

## Versiones que fija el DAS

| Herramienta | Versión |
|---|---|
| Go | `1.26.0` + `toolchain go1.26.11` |
| Huma | `v2.34.1` |
| Node.js | `22.22.0` |
| pnpm | `10.13.1` |
| goose | `v3.24.1` |
| sqlc | `v1.29.0` |
| oasdiff | `v1.35.0` |
| oapi-codegen | `v2.8.2` |
| lefthook | `2.1.10` |
| gitleaks | `8.30.1` |
| trivy | `v0.76.1` |
| dockle | `v0.4.15` |

## Arch / CachyOS

```bash
sudo pacman -Syu
sudo pacman -S --needed go git curl jq base-devel
sudo pacman -S --needed nodejs npm docker docker-compose docker-buildx
sudo pacman -S --needed gitleaks trivy
```

> `pacman` es atómico: si en una misma invocación incluyes un paquete que no
existe (por ejemplo `hadolint`), aborta **toda** la transacción y no instala
nada de lo demás. Instala en comandos separados y verifica los nombres antes.

`hadolint` no está en los repos oficiales:

```bash
mkdir -p "$HOME/.local/bin"
curl -fsSL -o "$HOME/.local/bin/hadolint" https://github.com/hadolint/hadolint/releases/download/v2.15.1/hadolint-linux-x86_64
chmod +x "$HOME/.local/bin/hadolint"
```

`lefthook` no se instala aparte: es `devDependency` del repo y se usa con
`pnpm exec` (ver [Hooks locales](#hooks-locales-lefthook--gitleaks)).

Si estás en ARM, cambia `hadolint-linux-x86_64` por `hadolint-linux-arm64`.

Alternativas AUR:

```bash
paru -S hadolint-bin lefthook dockle-bin
```

## Ubuntu / Debian

```bash
sudo apt update
sudo apt install -y git curl jq build-essential golang-go nodejs npm docker.io docker-compose-v2
```

Notas:

- `golang-go` de Ubuntu puede ser anterior a 1.26. El `go.mod` pide
  `toolchain go1.26.11`, así que Go intentará descargarla automáticamente. Si
  falla, instala la toolchain exacta:
  ```bash
  go install golang.org/dl/go1.26.11@latest
  "$HOME/go/bin/go1.26.11" download
  ```
  O baja el tarball desde `https://go.dev/dl/`.
- `nodejs` de Ubuntu puede ser anterior a 22.22.0. Para la versión exacta usa
  `fnm` o NodeSource:
  ```bash
  curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
  sudo apt install -y nodejs
  ```
- `docker-compose-v2` instala el comando `docker compose`. Si tu Ubuntu no lo
  tiene, usa el repositorio oficial de Docker o el paquete
  `docker-compose-plugin`.

### Herramientas que no vienen en apt

```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.24.1
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0
go install github.com/oasdiff/oasdiff@v1.35.0
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.2
go install github.com/gitleaks/gitleaks/v8@v8.30.1
go install github.com/aquasecurity/trivy/cmd/trivy@v0.76.1
```

`lefthook` no va en esta lista: es `devDependency` del repo y se resuelve con
`pnpm exec lefthook` después de `pnpm install`.

Hadolint:

```bash
mkdir -p "$HOME/.local/bin"
curl -fsSL -o "$HOME/.local/bin/hadolint" https://github.com/hadolint/hadolint/releases/download/v2.15.1/hadolint-linux-x86_64
chmod +x "$HOME/.local/bin/hadolint"
```

## PATH de Go y pnpm

`go install` deja los binarios en `$GOBIN` o, si no está definido, en
`$GOPATH/bin`. Verifica con:

```bash
go env GOPATH GOBIN
ls "$HOME/go/bin"
```

En **fish** (CachyOS suele usar fish), agrega el directorio al PATH:

```fish
fish_add_path "$HOME/go/bin"
fish_add_path "$HOME/.local/bin"
exec fish
```

En **bash/zsh**, agrégalo a `~/.bashrc` o `~/.zshrc`:

```bash
export PATH="$HOME/go/bin:$HOME/.local/bin:$PATH"
```

Comprueba:

```bash
which goose sqlc oasdiff oapi-codegen
goose --version
pnpm exec lefthook version
```

### pnpm con corepack

Si `corepack enable` falla con `EACCES` porque intenta escribir en
`/usr/bin/pnpm`, usa un directorio de usuario:

```bash
mkdir -p "$HOME/.local/bin"
corepack enable --install-directory "$HOME/.local/bin"
corepack prepare pnpm@10.13.1 --activate
pnpm --version
```

Alternativa:

```bash
corepack install -g pnpm@10.13.1
```

Si `pnpm` avisa `Unsupported engine: wanted 22.22.0 (current v26.12.1)`, es solo
un aviso: los scripts corren igual y el CI usa `.nvmrc` (22.22.0), así que no
afecta al pipeline. Para alinearlo en local, `fnm use` o `nvm use` dentro del
repo.

## Docker

```bash
sudo systemctl enable --now docker
sudo usermod -aG docker "$USER"
```

Cierra sesión y vuelve a entrar para que tome el grupo `docker`.

Dockle se usa como imagen, no se instala:

```bash
docker run --rm goodwithtech/dockle:v0.4.15 --version
```

El tag correcto lleva `v`. `goodwithtech/dockle:0.4.15` no existe.

## Docker Compose

```bash
cp .env.example .env        # completar POSTGRES_PASSWORD antes de levantar
docker compose up -d --build
docker compose ps
docker compose logs -f backend
docker compose down        # conserva el volumen de PostgreSQL
docker compose down -v     # borra también los datos
```

| Servicio | Puerto en el host | Notas |
|---|---|---|
| `backend` | `127.0.0.1:5131` → `8080` | `API_ENV=dev`, sirve `/api/docs` |
| `postgres` | `127.0.0.1:5433` → `5432` | volumen `postgres-datos` |

- Los puertos siguen la convención del DAS (`51N1` para la API, con N=3) y se
  publican solo en loopback. Se pueden cambiar con `API_PUBLISHED_PORT` y
  `POSTGRES_PUBLISHED_PORT`.
- `migrate`, `seed`, `frontend` y `mailpit` se agregan en el Hito 2 y en el
  ensamblaje LPS.
- `.env` es obligatorio: `POSTGRES_PASSWORD` no tiene valor por defecto, para no
dejar una credencial en el repositorio. Copia `.env.example` a `.env` y completa
los `replace-with-*`. `.env` no se commitea.
- La imagen de PostgreSQL está fijada por digest; no la cambies por `latest`.

## Hooks locales (lefthook + gitleaks)

`lefthook` es una `devDependency` (2.1.10): `pnpm install` lo deja en
`node_modules` y `pnpm exec` lo resuelve, sin instalarlo aparte ni tenerlo en el
PATH. `.npmrc` tiene `ignore-scripts=true`, así que los hooks no se instalan
solos:

```bash
pnpm install
pnpm exec lefthook install
ls -l .git/hooks/pre-commit .git/hooks/pre-push
```

Sin `pnpm exec lefthook install` no hay hooks. El `pre-commit` corre gofmt,
`go vet`, Prettier/oxlint (cuando exista `apps/web`) y Gitleaks; el `pre-push`
corre `pnpm verify`, `pnpm test:api` y verifica que el contrato esté al día.

Prueba que Gitleaks bloquea un secreto (el commit debe abortar):

```bash
TOKEN_FALSO="glpat-$(openssl rand -hex 10)"
printf 'token = %s\n' "$TOKEN_FALSO" > secreto-falso.txt
git add secreto-falso.txt
git commit -m "chore: prueba de hook"     # falla con leaks found: 1
git restore --staged secreto-falso.txt
rm -f secreto-falso.txt
```

El señuelo se genera al vuelo a propósito. Un `glpat-` literal pegado en la
documentación dispara Gitleaks **y** la protección de push de GitHub, que escanea
el contenido del commit y no respeta las anotaciones `gitleaks:allow`.

Para ver el detalle de un hallazgo, corre el mismo comando a mano:

```bash
gitleaks protect --staged --no-banner -v
```

No sirven como prueba las claves de ejemplo de la documentación de AWS: están en
la allowlist por defecto, y una cadena `AKIA` + 16 caracteres tampoco dispara
`aws-access-token` en 8.30.1. Con `glpat-` o `sk_live_` sí se ve el hallazgo.

## Verificación rápida

Desde `apps/api`:

```bash
go mod tidy
gofmt -l .
go vet ./...
go build ./...
go run ./cmd/codegen
go run ./cmd/api
```

El contrato queda en `contracts/openapi.yaml` y la API en
`http://localhost:8080`. La documentación interactiva está en
`http://localhost:8080/api/docs`.

Con Compose no hace falta `go run`: `docker compose up -d --build` deja la API en
`http://localhost:5131` y PostgreSQL en `localhost:5433`.

## Qué se versiona y qué no

**Sí se versiona:**

- Código fuente, `go.mod`, `go.sum` y `pnpm-lock.yaml`.
- `contracts/openapi.yaml`, aunque sea generado: el DAS pide commitearlo.
- `docs/`, `README.md` y `.env.example`.
- `db/migrations/`, `db/queries/` e `internal/repository/` generado por sqlc.
- `e2e/testdata/*.golden.md`.
- `Dockerfile`, `Dockerfile.migrate`, `docker-compose.yaml`, `.github/workflows/ci.yml`,
  `lefthook.yml`, `.npmrc` y `.nvmrc`.

**No se versiona:**

- `.env` y cualquier secreto (`*.pem`, `*.key`, etc.).
- Binarios, `coverage`, logs y archivos temporales.
- `node_modules/`, `dist/`, `build/`.
- Configuración local de editores (`.idea/`, `.vscode/`, `.zed/`) y archivos del sistema.
- `package-lock.json` y `yarn.lock`: el repo usa pnpm.
- `.scannerwork/`, `.trivy/` y reportes locales de escáneres.
- Datos locales de PostgreSQL (`pgdata/`, `.pgdata/`, archivos `.db`).

El material del curso vive en `docs/DAS — Desarrollo de Microservicios.md` y se
versiona con el resto de la documentación.
