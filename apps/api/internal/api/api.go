package api

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

// NuevaAPI crea la API Huma con la configuración del servicio de productos.
func NuevaAPI(mux *http.ServeMux, version, entorno string) huma.API {
	config := huma.DefaultConfig("Servicio de Productos", version)
	config.Info.Description = "CRUD de productos con SKU único, categorización múltiple, costo entero en pesos y flag de venta."
	config.OpenAPIPath = "/api/openapi.json"
	config.DocsPath = "/api/docs"

	if entorno == "prod" {
		config.OpenAPIPath = ""
		config.DocsPath = ""
	}

	return humago.New(mux, config)
}
