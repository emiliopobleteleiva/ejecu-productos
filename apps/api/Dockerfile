# Etapa 1: compilar
FROM golang:1.26.5-alpine3.23@sha256:622e56dbc11a8cfe87cafa2331e9a201877271cbff918af53d3be315f3da88cc AS build

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
FROM gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6

COPY --from=build /out/api /api

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]