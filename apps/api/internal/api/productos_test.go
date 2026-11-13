package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func nuevoMuxDePrueba(t *testing.T) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	api := NuevaAPI(mux, "test", "test")
	Registrar(api, NuevosManejadoresProductos())
	return mux
}

func ejecutar(t *testing.T, mux *http.ServeMux, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if cuerpo != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestProductos(t *testing.T) {
	mux := nuevoMuxDePrueba(t)

	t.Run("listado inicial", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodGet, "/api/productos", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("esperaba 200, obtuve %d: %s", rec.Code, rec.Body.String())
		}
		var body ListarProductosBody
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("respuesta inválida: %v", err)
		}
		if body.Total != 2 {
			t.Fatalf("esperaba 2 productos, obtuve %d", body.Total)
		}
	})

	var creado Producto

	t.Run("crear", func(t *testing.T) {
		cuerpo := `{"sku":"PROD-TEST","nombre":"Producto de prueba","descripcion":"Creado por el test","categoria_ids":["7c9e6679-7425-40de-944b-e07fc1f90ae7"],"costo":1500,"permite_venta":true}`
		rec := ejecutar(t, mux, http.MethodPost, "/api/productos", cuerpo)
		if rec.Code != http.StatusCreated {
			t.Fatalf("esperaba 201, obtuve %d: %s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get("Location") == "" {
			t.Fatal("falta el header Location")
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &creado); err != nil {
			t.Fatalf("respuesta inválida: %v", err)
		}
		if creado.ID == "" || creado.SKU != "PROD-TEST" {
			t.Fatalf("producto inesperado: %+v", creado)
		}
	})

	t.Run("obtener", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodGet, "/api/productos/"+creado.ID, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("esperaba 200, obtuve %d", rec.Code)
		}
	})

	t.Run("listar por categoria", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodGet, "/api/productos?categoria_id=7c9e6679-7425-40de-944b-e07fc1f90ae7", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("esperaba 200, obtuve %d", rec.Code)
		}
		var body ListarProductosBody
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("respuesta inválida: %v", err)
		}
		if body.Total < 2 {
			t.Fatalf("esperaba al menos 2 productos, obtuve %d", body.Total)
		}
	})

	t.Run("listar por permite_venta", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodGet, "/api/productos?permite_venta=false", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("esperaba 200, obtuve %d", rec.Code)
		}
		var body ListarProductosBody
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("respuesta inválida: %v", err)
		}
		for _, p := range body.Productos {
			if p.PermiteVenta {
				t.Fatalf("producto %s no debería permitir venta", p.ID)
			}
		}
	})

	t.Run("actualizar", func(t *testing.T) {
		cuerpo := `{"nombre":"Producto actualizado","costo":2000,"permite_venta":false}`
		rec := ejecutar(t, mux, http.MethodPatch, "/api/productos/"+creado.ID, cuerpo)
		if rec.Code != http.StatusOK {
			t.Fatalf("esperaba 200, obtuve %d: %s", rec.Code, rec.Body.String())
		}
		var actualizado Producto
		if err := json.Unmarshal(rec.Body.Bytes(), &actualizado); err != nil {
			t.Fatalf("respuesta inválida: %v", err)
		}
		if actualizado.Nombre != "Producto actualizado" || actualizado.Costo != 2000 || actualizado.PermiteVenta {
			t.Fatalf("producto inesperado: %+v", actualizado)
		}
	})

	t.Run("sku duplicado al crear", func(t *testing.T) {
		cuerpo := `{"sku":"PROD-001","nombre":"Duplicado","categoria_ids":["7c9e6679-7425-40de-944b-e07fc1f90ae7"],"costo":1000,"permite_venta":true}`
		rec := ejecutar(t, mux, http.MethodPost, "/api/productos", cuerpo)
		if rec.Code != http.StatusConflict {
			t.Fatalf("esperaba 409, obtuve %d", rec.Code)
		}
	})

	t.Run("sku duplicado al actualizar", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodPatch, "/api/productos/3fa85f64-5717-4562-b3fc-2c963f66afa6", `{"sku":"PROD-002"}`)
		if rec.Code != http.StatusConflict {
			t.Fatalf("esperaba 409, obtuve %d", rec.Code)
		}
	})

	validaciones := []struct {
		nombre string
		cuerpo string
	}{
		{"categorias vacias", `{"sku":"PROD-V1","nombre":"V1","categoria_ids":[],"costo":1000,"permite_venta":true}`},
		{"categorias duplicadas", `{"sku":"PROD-V2","nombre":"V2","categoria_ids":["7c9e6679-7425-40de-944b-e07fc1f90ae7","7c9e6679-7425-40de-944b-e07fc1f90ae7"],"costo":1000,"permite_venta":true}`},
		{"categoria uuid invalido", `{"sku":"PROD-V3","nombre":"V3","categoria_ids":["no-es-uuid"],"costo":1000,"permite_venta":true}`},
		{"categoria faltante", `{"sku":"PROD-V4","nombre":"V4","costo":1000,"permite_venta":true}`},
		{"costo negativo", `{"sku":"PROD-V5","nombre":"V5","categoria_ids":["7c9e6679-7425-40de-944b-e07fc1f90ae7"],"costo":-1,"permite_venta":true}`},
		{"costo decimal", `{"sku":"PROD-V6","nombre":"V6","categoria_ids":["7c9e6679-7425-40de-944b-e07fc1f90ae7"],"costo":1.5,"permite_venta":true}`},
		{"sku invalido", `{"sku":"prod 1","nombre":"V7","categoria_ids":["7c9e6679-7425-40de-944b-e07fc1f90ae7"],"costo":1000,"permite_venta":true}`},
	}
	for _, caso := range validaciones {
		t.Run("validacion "+caso.nombre, func(t *testing.T) {
			rec := ejecutar(t, mux, http.MethodPost, "/api/productos", caso.cuerpo)
			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("esperaba 422, obtuve %d: %s", rec.Code, rec.Body.String())
			}
		})
	}

	t.Run("uuid invalido en path", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodGet, "/api/productos/no-es-uuid", "")
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("esperaba 422, obtuve %d", rec.Code)
		}
	})

	t.Run("eliminar", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodDelete, "/api/productos/"+creado.ID, "")
		if rec.Code != http.StatusNoContent {
			t.Fatalf("esperaba 204, obtuve %d", rec.Code)
		}
		rec = ejecutar(t, mux, http.MethodGet, "/api/productos/"+creado.ID, "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("esperaba 404, obtuve %d", rec.Code)
		}
		rec = ejecutar(t, mux, http.MethodDelete, "/api/productos/"+creado.ID, "")
		if rec.Code != http.StatusNoContent {
			t.Fatalf("esperaba 204 idempotente, obtuve %d", rec.Code)
		}
	})

	t.Run("obtener inexistente", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodGet, "/api/productos/00000000-0000-4000-8000-000000000000", "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("esperaba 404, obtuve %d", rec.Code)
		}
	})

	t.Run("eliminar inexistente", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodDelete, "/api/productos/00000000-0000-4000-8000-000000000000", "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("esperaba 404, obtuve %d", rec.Code)
		}
	})
}
