package api

import (
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// Producto representa un producto del catálogo.
//
// La categorización es múltiple: un producto puede pertenecer a una o más
// categorías. En la base de datos del Hito 2 se modelará con la tabla puente
// producto_categorias(producto_id, categoria_id).
type Producto struct {
	ID            string    `json:"id" format:"uuid" readOnly:"true" doc:"Identificador único del producto." example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	SKU           string    `json:"sku" minLength:"3" maxLength:"64" pattern:"^[A-Z0-9][A-Z0-9_-]{2,63}$" doc:"SKU único del producto." example:"PROD-001"`
	Nombre        string    `json:"nombre" minLength:"1" maxLength:"200" doc:"Nombre visible del producto." example:"Café molido 1 kg"`
	Descripcion   string    `json:"descripcion,omitempty" maxLength:"1000" doc:"Descripción libre del producto." example:"Café tostado de origen."`
	CategoriaIDs  []string  `json:"categoria_ids" nullable:"false" minItems:"1" uniqueItems:"true" doc:"Identificadores UUID de las categorías a las que pertenece el producto." example:"7c9e6679-7425-40de-944b-e07fc1f90ae7,16fd2706-8baf-433b-82eb-8c7fada847da"`
	Costo         int64     `json:"costo" minimum:"0" doc:"Costo en pesos chilenos, entero." example:"12990"`
	PermiteVenta  bool      `json:"permite_venta" doc:"Indica si el producto puede venderse." example:"true"`
	Activo        bool      `json:"activo" readOnly:"true" doc:"Indica si el producto está activo." example:"true"`
	CreadoEn      time.Time `json:"creado_en" readOnly:"true" doc:"Fecha de creación (timestamptz en Hito 2)."`
	ActualizadoEn time.Time `json:"actualizado_en" readOnly:"true" doc:"Fecha de última actualización (timestamptz en Hito 2)."`
}

// TransformSchema marca cada elemento de categoria_ids como UUID. Huma v2 no
// permite expresar el formato de los ítems de un []string con un tag, así que
// se ajusta el esquema generado sin cambiar el tipo Go del campo.
func (Producto) TransformSchema(_ huma.Registry, s *huma.Schema) *huma.Schema {
	marcarCategoriaIDsComoUUID(s)
	return s
}

// CrearProductoBody es el cuerpo de la creación de un producto.
type CrearProductoBody struct {
	SKU          string   `json:"sku" required:"true" minLength:"3" maxLength:"64" pattern:"^[A-Z0-9][A-Z0-9_-]{2,63}$" doc:"SKU único del producto." example:"PROD-001"`
	Nombre       string   `json:"nombre" required:"true" minLength:"1" maxLength:"200" doc:"Nombre visible del producto." example:"Café molido 1 kg"`
	Descripcion  string   `json:"descripcion,omitempty" maxLength:"1000" doc:"Descripción libre del producto." example:"Café tostado de origen."`
	CategoriaIDs []string `json:"categoria_ids" required:"true" nullable:"false" minItems:"1" uniqueItems:"true" doc:"Identificadores UUID de las categorías. Debe incluir al menos una." example:"7c9e6679-7425-40de-944b-e07fc1f90ae7,16fd2706-8baf-433b-82eb-8c7fada847da"`
	Costo        int64    `json:"costo" required:"true" minimum:"0" doc:"Costo en pesos chilenos, entero." example:"12990"`
	PermiteVenta bool     `json:"permite_venta" required:"true" doc:"Indica si el producto puede venderse." example:"true"`
}

func (CrearProductoBody) TransformSchema(_ huma.Registry, s *huma.Schema) *huma.Schema {
	marcarCategoriaIDsComoUUID(s)
	return s
}

// CrearProductoInput es la entrada de POST /api/productos.
type CrearProductoInput struct {
	Body CrearProductoBody
}

// CrearProductoOutput es la salida de POST /api/productos.
type CrearProductoOutput struct {
	Location string `header:"Location" doc:"URL del producto creado."`
	Body     Producto
}

// ListarProductosInput es la entrada de GET /api/productos.
type ListarProductosInput struct {
	CategoriaID  string                  `query:"categoria_id" format:"uuid" doc:"Filtra por una categoría (UUID). Es solo un filtro, no un campo del producto." example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	PermiteVenta ParametroOpcional[bool] `query:"permite_venta" doc:"Filtra por disponibilidad de venta." example:"true"`
}

// ListarProductosBody es el cuerpo de la respuesta del listado.
type ListarProductosBody struct {
	Productos []Producto `json:"productos" nullable:"false" doc:"Productos que cumplen los filtros."`
	Total     int        `json:"total" doc:"Cantidad total de productos que cumplen los filtros." example:"2"`
}

// ListarProductosOutput es la salida de GET /api/productos.
type ListarProductosOutput struct {
	Body ListarProductosBody
}

// ObtenerProductoInput es la entrada de GET /api/productos/{id}.
type ObtenerProductoInput struct {
	ID string `path:"id" format:"uuid" doc:"Identificador UUID del producto."`
}

// ObtenerProductoOutput es la salida de GET /api/productos/{id}.
type ObtenerProductoOutput struct {
	Body Producto
}

// ActualizarProductoBody es el cuerpo de PATCH /api/productos/{id}. Los
// punteros permiten distinguir "campo no enviado" de "campo enviado con valor
// cero".
type ActualizarProductoBody struct {
	SKU          *string   `json:"sku,omitempty" minLength:"3" maxLength:"64" pattern:"^[A-Z0-9][A-Z0-9_-]{2,63}$" doc:"Nuevo SKU único del producto." example:"PROD-001"`
	Nombre       *string   `json:"nombre,omitempty" minLength:"1" maxLength:"200" doc:"Nuevo nombre visible del producto." example:"Café molido 1 kg"`
	Descripcion  *string   `json:"descripcion,omitempty" maxLength:"1000" doc:"Nueva descripción libre del producto." example:"Café tostado de origen."`
	CategoriaIDs *[]string `json:"categoria_ids,omitempty" nullable:"false" minItems:"1" uniqueItems:"true" doc:"Nuevos identificadores UUID de categorías. Si se envía, debe incluir al menos una." example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	Costo        *int64    `json:"costo,omitempty" minimum:"0" doc:"Nuevo costo en pesos chilenos, entero." example:"12990"`
	PermiteVenta *bool     `json:"permite_venta,omitempty" doc:"Nuevo valor del flag de venta." example:"true"`
}

func (ActualizarProductoBody) TransformSchema(_ huma.Registry, s *huma.Schema) *huma.Schema {
	marcarCategoriaIDsComoUUID(s)
	return s
}

// ActualizarProductoInput es la entrada de PATCH /api/productos/{id}.
type ActualizarProductoInput struct {
	ID   string `path:"id" format:"uuid" doc:"Identificador UUID del producto."`
	Body ActualizarProductoBody
}

// ActualizarProductoOutput es la salida de PATCH /api/productos/{id}.
type ActualizarProductoOutput struct {
	Body Producto
}

// EliminarProductoInput es la entrada de DELETE /api/productos/{id}.
type EliminarProductoInput struct {
	ID string `path:"id" format:"uuid" doc:"Identificador UUID del producto."`
}

// ParametroOpcional permite distinguir "no enviado" de "enviado con valor
// cero" en un query param. Huma v2 no admite punteros en parámetros de ruta,
// query o header, así que se usa este wrapper.
type ParametroOpcional[T any] struct {
	Valor    T
	Definido bool
}

// Schema expone el esquema del valor envuelto.
func (p ParametroOpcional[T]) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeFor[T]())
}

// Receiver devuelve el campo donde Huma debe escribir el valor parseado.
func (p *ParametroOpcional[T]) Receiver() reflect.Value {
	return reflect.ValueOf(p).Elem().Field(0)
}

// OnParamSet registra si el parámetro vino o no en la request.
func (p *ParametroOpcional[T]) OnParamSet(definido bool, _ any) {
	p.Definido = definido
}

// marcarCategoriaIDsComoUUID ajusta el esquema generado para que los ítems de
// categoria_ids declaren format: uuid.
func marcarCategoriaIDsComoUUID(s *huma.Schema) {
	if s == nil || s.Properties == nil {
		return
	}
	prop, ok := s.Properties["categoria_ids"]
	if !ok || prop == nil || prop.Items == nil {
		return
	}
	prop.Items.Format = "uuid"
}
