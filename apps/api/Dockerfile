# Etapa 1: compilar
FROM golang:1.26.11-alpine3.23@sha256:db99415e2625ce0ee03eed09edf7f32ae8c58dc399f28cc602a87fe7238119e4 AS build

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