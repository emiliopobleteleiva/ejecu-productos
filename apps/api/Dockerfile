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
FROM gcr.io/distroless/static-debian13:nonroot@sha256:aafe993665bf6200db38a02c11a9351e782a84b6a45e7f33b6695638c88e3931

COPY --from=build /out/api /api

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
