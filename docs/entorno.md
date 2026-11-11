# Entorno de desarrollo

Requisitos para levantar el servicio y ejecutar las herramientas del DAS.
El repo se probó en CachyOS/Arch; los comandos para Ubuntu/Debian se incluyen
para el resto del grupo.

## Versiones que fija el DAS

| Herramienta | Versión |
|---|---|
| Go | `1.26.0` + `toolchain go1.26.5` |
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

`hadolint` y `lefthook` no están en los repos oficiales:

```bash
mkdir -p "$HOME/.local/bin"
curl -fsSL -o "$HOME/.local/bin/hadolint" https://github.com/hadolint/hadolint/releases/download/v2.15.1/hadolint-linux-x86_64
chmod +x "$HOME/.local/bin/hadolint"

go install github.com/evilmartians/lefthook/v2@v2.1.10
```

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
  `toolchain go1.26.5`, así que Go intentará descargarla automáticamente. Si
  falla, instala la toolchain exacta:
  ```bash
  go install golang.org/dl/go1.26.5@latest
  "$HOME/go/bin/go1.26.5" download
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
go install github.com/evilmartians/lefthook/v2@v2.1.10
go install github.com/gitleaks/gitleaks/v8@v8.30.1
go install github.com/aquasecurity/trivy/cmd/trivy@v0.76.1
```

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
which goose sqlc oasdiff oapi-codegen lefthook
goose --version
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

`enunciado/` no está ignorado por defecto. Si no quieres versionar el material
del curso, descomenta `/enunciado/` en `.gitignore`. Si quieres conservar el DAS
en el repo, déjalo como está y commitéalo.
