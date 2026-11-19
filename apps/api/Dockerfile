# Etapa 1: compilar
FROM golang:1.27.4-alpine3.23@sha256:f70850ef6ce341e8b6427f677049145d63c0b7ce002d4d7e3d464828b1b96dcd AS build

ENV GOTOOLCHAIN=local
WORKDIR /src

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download && go mod verify

COPY apps/api/ ./

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -buildvcs=false \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/api ./cmd/api

# Etapa 2: ejecutar
FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329

COPY --from=build /out/api /api

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
