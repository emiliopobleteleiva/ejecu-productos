package api

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

// NuevaAPI crea la API Huma con la configuración del servicio de productos.
func NuevaAPI(mux *http.ServeMux, version, entorno string) huma.API {
	config := huma.DefaultConfig("Servicio de Productos", version)
	config.Info.Description = "CRUD de productos con SKU autogenerado e inmutable por departamento, categorización múltiple, costo unitario con IVA incluido y flag de venta."
	// Huma agrega la extensión: publica <ruta>.json y <ruta>.yaml. Sin extensión
	// en la configuración, el spec queda en /api/openapi.json y /api/openapi.yaml.
	config.OpenAPIPath = "/api/openapi"
	config.DocsPath = "/api/docs"

	if entorno == "prod" {
		config.OpenAPIPath = ""
		config.DocsPath = ""
	}

	return humago.New(mux, config)
}
