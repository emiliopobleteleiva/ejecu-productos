package api

import (
	"context"
	"sort"
)

// Departamento representa un departamento del catálogo.
//
// En el Hito 2 vive en su propia tabla; en el Hito 1 es un catálogo simulado.
// Se publica para que el front y los servicios consumidores sepan qué
// departamento_id enviar al crear un producto.
type Departamento struct {
	ID     string `json:"id" format:"uuid" readOnly:"true" doc:"Identificador UUID del departamento." example:"1b4e28ba-2fa1-11d2-883f-0016d3cca427"`
	Codigo string `json:"codigo" minLength:"2" maxLength:"8" pattern:"^[A-Z0-9]{2,8}$" doc:"Código que se usa como prefijo del SKU de sus productos." example:"BEB"`
	Nombre string `json:"nombre" minLength:"1" maxLength:"120" doc:"Nombre visible del departamento." example:"Bebidas"`
}

// departamentosDemo simula la tabla de departamentos del Hito 2.
var departamentosDemo = []Departamento{
	{ID: "1b4e28ba-2fa1-11d2-883f-0016d3cca427", Codigo: "BEB", Nombre: "Bebidas"},
	{ID: "2c4e28ba-2fa1-11d2-883f-0016d3cca428", Codigo: "HOG", Nombre: "Hogar"},
}

// buscarDepartamento devuelve el departamento por su UUID.
func buscarDepartamento(id string) (Departamento, bool) {
	for _, d := range departamentosDemo {
		if d.ID == id {
			return d, true
		}
	}
	return Departamento{}, false
}

// ListarDepartamentosInput es la entrada de GET /api/departamentos.
type ListarDepartamentosInput struct{}

// ListarDepartamentosBody es el cuerpo de la respuesta del catálogo.
type ListarDepartamentosBody struct {
	Departamentos []Departamento `json:"departamentos" nullable:"false" doc:"Departamentos disponibles."`
	Total         int            `json:"total" doc:"Cantidad de departamentos." example:"2"`
}

// ListarDepartamentosOutput es la salida de GET /api/departamentos.
type ListarDepartamentosOutput struct {
	Body ListarDepartamentosBody
}

// listarDepartamentos implementa GET /api/departamentos.
//
// TODO Hito 2: leer el catálogo desde la tabla departamentos.
func (*ManejadoresProductos) listarDepartamentos(_ context.Context, _ *ListarDepartamentosInput) (*ListarDepartamentosOutput, error) {
	departamentos := append([]Departamento(nil), departamentosDemo...)
	sort.Slice(departamentos, func(i, j int) bool {
		return departamentos[i].Codigo < departamentos[j].Codigo
	})
	return &ListarDepartamentosOutput{
		Body: ListarDepartamentosBody{
			Departamentos: departamentos,
			Total:         len(departamentos),
		},
	}, nil
}
