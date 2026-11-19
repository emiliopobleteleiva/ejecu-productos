package api

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// Registrar registra las operaciones de productos en la API Huma.
func Registrar(api huma.API, manejadores *ManejadoresProductos) {
	huma.Register(api, huma.Operation{
		OperationID:   "crear-producto",
		Method:        http.MethodPost,
		Path:          "/api/productos",
		Summary:       "Crea un producto",
		Description:   "Crea un producto con SKU autogenerado a partir del departamento y al menos una categoría. El SKU es inmutable.",
		Tags:          []string{"productos"},
		DefaultStatus: http.StatusCreated,
	}, manejadores.crear)

	huma.Register(api, huma.Operation{
		OperationID: "listar-productos",
		Method:      http.MethodGet,
		Path:        "/api/productos",
		Summary:     "Lista productos",
		Description: "Lista los productos activos y permite filtrar por categoría, por departamento y por disponibilidad de venta.",
		Tags:        []string{"productos"},
	}, manejadores.listar)

	huma.Register(api, huma.Operation{
		OperationID: "listar-departamentos",
		Method:      http.MethodGet,
		Path:        "/api/departamentos",
		Summary:     "Lista los departamentos disponibles",
		Description: "Devuelve el catálogo de departamentos con el id que se envía al crear un producto y el código que forma su SKU.",
		Tags:        []string{"departamentos"},
	}, manejadores.listarDepartamentos)

	huma.Register(api, huma.Operation{
		OperationID: "obtener-producto",
		Method:      http.MethodGet,
		Path:        "/api/productos/{id}",
		Summary:     "Obtiene un producto por su identificador",
		Description: "Devuelve un producto activo por su UUID.",
		Tags:        []string{"productos"},
	}, manejadores.obtener)

	huma.Register(api, huma.Operation{
		OperationID: "actualizar-producto",
		Method:      http.MethodPatch,
		Path:        "/api/productos/{id}",
		Summary:     "Actualiza parcialmente un producto",
		Description: "Actualiza solo los campos enviados. El SKU y el departamento son inmutables.",
		Tags:        []string{"productos"},
	}, manejadores.actualizar)

	huma.Register(api, huma.Operation{
		OperationID:   "eliminar-producto",
		Method:        http.MethodDelete,
		Path:          "/api/productos/{id}",
		Summary:       "Desactiva un producto",
		Description:   "Aplica una baja lógica: el producto deja de listarse y de ser accesible por ID.",
		Tags:          []string{"productos"},
		DefaultStatus: http.StatusNoContent,
	}, manejadores.eliminar)
}
