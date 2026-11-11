package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// ManejadoresProductos implementa los handlers del Hito 1 con datos en
// memoria. Esta estructura y su almacenamiento se reemplazan por service y
// store en el Hito 2.
type ManejadoresProductos struct {
	mu        sync.RWMutex
	productos map[string]*Producto
	skus      map[string]string
}

// NuevosManejadoresProductos crea el almacén simulado con datos de demo.
func NuevosManejadoresProductos() *ManejadoresProductos {
	ahora := time.Date(2026, time.November, 11, 0, 0, 0, 0, time.UTC)

	productos := []Producto{
		{
			ID:            "3fa85f64-5717-4562-b3fc-2c963f66afa6",
			SKU:           "PROD-001",
			Nombre:        "Café molido 1 kg",
			Descripcion:   "Café tostado de origen.",
			CategoriaIDs:  []string{"7c9e6679-7425-40de-944b-e07fc1f90ae7"},
			Costo:         12990,
			PermiteVenta:  true,
			Activo:        true,
			CreadoEn:      ahora,
			ActualizadoEn: ahora,
		},
		{
			ID:            "9b2b6c1e-4f3a-4c1d-9a7e-2f5b8c3d1e4f",
			SKU:           "PROD-002",
			Nombre:        "Taza cerámica",
			Descripcion:   "Taza de 350 ml.",
			CategoriaIDs:  []string{"16fd2706-8baf-433b-82eb-8c7fada847da"},
			Costo:         4990,
			PermiteVenta:  false,
			Activo:        true,
			CreadoEn:      ahora,
			ActualizadoEn: ahora,
		},
	}

	m := &ManejadoresProductos{
		productos: make(map[string]*Producto, len(productos)),
		skus:      make(map[string]string, len(productos)),
	}
	for i := range productos {
		p := productos[i]
		m.productos[p.ID] = &p
		m.skus[p.SKU] = p.ID
	}
	return m
}

// crear implementa POST /api/productos.
func (m *ManejadoresProductos) crear(_ context.Context, input *CrearProductoInput) (*CrearProductoOutput, error) {
	// TODO: permiso productos:administrar
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, existe := m.skus[input.Body.SKU]; existe {
		return nil, huma.Error409Conflict("Ya existe un producto con el SKU indicado.")
	}

	ahora := time.Now().UTC()
	producto := &Producto{
		ID:            nuevoUUID(),
		SKU:           input.Body.SKU,
		Nombre:        input.Body.Nombre,
		Descripcion:   input.Body.Descripcion,
		CategoriaIDs:  append([]string(nil), input.Body.CategoriaIDs...),
		Costo:         input.Body.Costo,
		PermiteVenta:  input.Body.PermiteVenta,
		Activo:        true,
		CreadoEn:      ahora,
		ActualizadoEn: ahora,
	}

	m.productos[producto.ID] = producto
	m.skus[producto.SKU] = producto.ID

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

// actualizar implementa PATCH /api/productos/{id}.
func (m *ManejadoresProductos) actualizar(_ context.Context, input *ActualizarProductoInput) (*ActualizarProductoOutput, error) {
	// TODO: permiso productos:administrar
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.productos[input.ID]
	if !ok || !p.Activo {
		return nil, huma.Error404NotFound("Producto no encontrado.")
	}

	if input.Body.SKU != nil && *input.Body.SKU != p.SKU {
		if id, existe := m.skus[*input.Body.SKU]; existe && id != p.ID {
			return nil, huma.Error409Conflict("Ya existe un producto con el SKU indicado.")
		}
		delete(m.skus, p.SKU)
		p.SKU = *input.Body.SKU
		m.skus[p.SKU] = p.ID
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
	if input.Body.Costo != nil {
		p.Costo = *input.Body.Costo
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
