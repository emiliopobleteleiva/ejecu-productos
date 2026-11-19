package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// departamentosDemo simula la tabla de departamentos del Hito 2. El código es
// un dato del departamento, no se deriva del UUID: acá se mantiene a mano y en
// el Hito 2 se resuelve desde la base o desde el servicio dueño del dato.
var departamentosDemo = map[string]string{
	"1b4e28ba-2fa1-11d2-883f-0016d3cca427": "BEB",
	"2c4e28ba-2fa1-11d2-883f-0016d3cca428": "HOG",
}

// ManejadoresProductos implementa los handlers del Hito 1 con datos en
// memoria. Esta estructura y su almacenamiento se reemplazan por service y
// store en el Hito 2.
type ManejadoresProductos struct {
	mu         sync.RWMutex
	productos  map[string]*Producto
	secuencias map[string]int
}

// NuevosManejadoresProductos crea el almacén simulado con datos de demo.
func NuevosManejadoresProductos() *ManejadoresProductos {
	ahora := time.Date(2026, time.November, 11, 0, 0, 0, 0, time.UTC)

	productos := []Producto{
		{
			ID:             "3fa85f64-5717-4562-b3fc-2c963f66afa6",
			SKU:            "BEB-000001",
			DepartamentoID: "1b4e28ba-2fa1-11d2-883f-0016d3cca427",
			Nombre:         "Café molido 1 kg",
			Descripcion:    "Café tostado de origen.",
			CategoriaIDs:   []string{"7c9e6679-7425-40de-944b-e07fc1f90ae7"},
			CostoUnitario:  12990,
			TipoImpuesto:   "IVA",
			PermiteVenta:   true,
			Activo:         true,
			CreadoEn:       ahora,
			ActualizadoEn:  ahora,
		},
		{
			ID:             "9b2b6c1e-4f3a-4c1d-9a7e-2f5b8c3d1e4f",
			SKU:            "HOG-000001",
			DepartamentoID: "2c4e28ba-2fa1-11d2-883f-0016d3cca428",
			Nombre:         "Taza cerámica",
			Descripcion:    "Taza de 350 ml.",
			CategoriaIDs:   []string{"16fd2706-8baf-433b-82eb-8c7fada847da"},
			CostoUnitario:  4990,
			TipoImpuesto:   "IVA",
			PermiteVenta:   false,
			Activo:         true,
			CreadoEn:       ahora,
			ActualizadoEn:  ahora,
		},
	}

	m := &ManejadoresProductos{
		productos:  make(map[string]*Producto, len(productos)),
		secuencias: make(map[string]int, len(departamentosDemo)),
	}
	for i := range productos {
		p := productos[i]
		m.productos[p.ID] = &p
		m.secuencias[p.DepartamentoID]++
	}
	return m
}

// crear implementa POST /api/productos. El SKU lo genera el servicio a partir
// del departamento; el cliente no lo envía.
func (m *ManejadoresProductos) crear(_ context.Context, input *CrearProductoInput) (*CrearProductoOutput, error) {
	// TODO: permiso productos:administrar
	// TODO Hito 2: validar que el departamento exista y generar el SKU dentro
	// de la misma transacción que inserta el producto.
	m.mu.Lock()
	defer m.mu.Unlock()

	ahora := time.Now().UTC()
	producto := &Producto{
		ID:             nuevoUUID(),
		SKU:            m.nuevoSKU(input.Body.DepartamentoID),
		DepartamentoID: input.Body.DepartamentoID,
		Nombre:         input.Body.Nombre,
		Descripcion:    input.Body.Descripcion,
		CategoriaIDs:   append([]string(nil), input.Body.CategoriaIDs...),
		CostoUnitario:  input.Body.CostoUnitario,
		TipoImpuesto:   input.Body.TipoImpuesto,
		PermiteVenta:   input.Body.PermiteVenta,
		Activo:         true,
		CreadoEn:       ahora,
		ActualizadoEn:  ahora,
	}

	m.productos[producto.ID] = producto

	return &CrearProductoOutput{
		Location: "/api/productos/" + producto.ID,
		Body:     clonarProducto(producto),
	}, nil
}

// listar implementa GET /api/productos.
func (m *ManejadoresProductos) listar(_ context.Context, input *ListarProductosInput) (*ListarProductosOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	productos := make([]Producto, 0, len(m.productos))
	for _, p := range m.productos {
		if !p.Activo {
			continue
		}
		if input.CategoriaID != "" && !contieneCategoria(p.CategoriaIDs, input.CategoriaID) {
			continue
		}
		if input.DepartamentoID != "" && p.DepartamentoID != input.DepartamentoID {
			continue
		}
		if input.PermiteVenta.Definido && p.PermiteVenta != input.PermiteVenta.Valor {
			continue
		}
		productos = append(productos, clonarProducto(p))
	}
	sort.Slice(productos, func(i, j int) bool {
		return productos[i].SKU < productos[j].SKU
	})

	return &ListarProductosOutput{
		Body: ListarProductosBody{
			Productos: productos,
			Total:     len(productos),
		},
	}, nil
}

// obtener implementa GET /api/productos/{id}.
func (m *ManejadoresProductos) obtener(_ context.Context, input *ObtenerProductoInput) (*ObtenerProductoOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.productos[input.ID]
	if !ok || !p.Activo {
		return nil, huma.Error404NotFound("Producto no encontrado.")
	}
	return &ObtenerProductoOutput{Body: clonarProducto(p)}, nil
}

// actualizar implementa PATCH /api/productos/{id}. El SKU y el departamento no
// se modifican: cambiarlos alteraría el identificador de negocio del producto.
func (m *ManejadoresProductos) actualizar(_ context.Context, input *ActualizarProductoInput) (*ActualizarProductoOutput, error) {
	// TODO: permiso productos:administrar
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.productos[input.ID]
	if !ok || !p.Activo {
		return nil, huma.Error404NotFound("Producto no encontrado.")
	}

	if input.Body.Nombre != nil {
		p.Nombre = *input.Body.Nombre
	}
	if input.Body.Descripcion != nil {
		p.Descripcion = *input.Body.Descripcion
	}
	if input.Body.CategoriaIDs != nil {
		p.CategoriaIDs = append([]string(nil), (*input.Body.CategoriaIDs)...)
	}
	if input.Body.CostoUnitario != nil {
		p.CostoUnitario = *input.Body.CostoUnitario
	}
	if input.Body.TipoImpuesto != nil {
		p.TipoImpuesto = *input.Body.TipoImpuesto
	}
	if input.Body.PermiteVenta != nil {
		p.PermiteVenta = *input.Body.PermiteVenta
	}
	p.ActualizadoEn = time.Now().UTC()

	return &ActualizarProductoOutput{Body: clonarProducto(p)}, nil
}

// eliminar implementa DELETE /api/productos/{id} como baja lógica.
func (m *ManejadoresProductos) eliminar(_ context.Context, input *EliminarProductoInput) (*struct{}, error) {
	// TODO: permiso productos:administrar
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.productos[input.ID]
	if !ok {
		return nil, huma.Error404NotFound("Producto no encontrado.")
	}
	if p.Activo {
		p.Activo = false
		p.ActualizadoEn = time.Now().UTC()
	}
	return nil, nil
}

// nuevoSKU arma el SKU del producto y avanza el secuencial del departamento.
// Debe llamarse con el mutex tomado. En el Hito 2 el secuencial sale de una
// secuencia por departamento dentro de la misma transacción del insert.
func (m *ManejadoresProductos) nuevoSKU(departamentoID string) string {
	m.secuencias[departamentoID]++
	return fmt.Sprintf("%s-%06d", codigoDepartamento(departamentoID), m.secuencias[departamentoID])
}

// codigoDepartamento devuelve el prefijo del SKU. Con un departamento conocido
// usa su código; con uno desconocido cae al prefijo del UUID para no dejar el
// producto sin SKU.
func codigoDepartamento(departamentoID string) string {
	if codigo, ok := departamentosDemo[departamentoID]; ok {
		return codigo
	}
	limpio := strings.ToUpper(strings.ReplaceAll(departamentoID, "-", ""))
	if len(limpio) < 4 {
		return "GEN"
	}
	return limpio[:4]
}

// clonarProducto devuelve una copia segura para no exponer los slices
// internos fuera del mutex.
func clonarProducto(p *Producto) Producto {
	copia := *p
	copia.CategoriaIDs = append([]string(nil), p.CategoriaIDs...)
	return copia
}

// contieneCategoria informa si el producto pertenece a la categoría indicada.
func contieneCategoria(categorias []string, categoriaID string) bool {
	for _, categoria := range categorias {
		if categoria == categoriaID {
			return true
		}
	}
	return false
}

// nuevoUUID genera un UUID v4 usando solo la biblioteca estándar.
func nuevoUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("no se pudo generar UUID: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}
