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

const (
	departamentoBebidas = "1b4e28ba-2fa1-11d2-883f-0016d3cca427"
	departamentoHogar   = "2c4e28ba-2fa1-11d2-883f-0016d3cca428"
	categoriaBebidas    = "7c9e6679-7425-40de-944b-e07fc1f90ae7"
)

// cuerpoCreacion arma un cuerpo válido de creación reemplazando el
// departamento y el campo que interese variar en cada caso.
func cuerpoCreacion(departamento, extra string) string {
	base := `{"departamento_id":"` + departamento + `","nombre":"Producto de prueba",` +
		`"categoria_ids":["` + categoriaBebidas + `"],"costo_unitario":1500,` +
		`"tipo_impuesto":"IVA","permite_venta":true`
	if extra != "" {
		base += "," + extra
	}
	return base + "}"
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

	t.Run("crear genera el sku desde el departamento", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodPost, "/api/productos", cuerpoCreacion(departamentoBebidas, ""))
		if rec.Code != http.StatusCreated {
			t.Fatalf("esperaba 201, obtuve %d: %s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get("Location") == "" {
			t.Fatal("falta el header Location")
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &creado); err != nil {
			t.Fatalf("respuesta inválida: %v", err)
		}
		if creado.ID == "" {
			t.Fatalf("falta el id: %+v", creado)
		}
		// El demo ya consumió el secuencial 1 de BEB.
		if creado.SKU != "BEB-000002" {
			t.Fatalf("esperaba BEB-000002, obtuve %q", creado.SKU)
		}
		if creado.DepartamentoID != departamentoBebidas {
			t.Fatalf("departamento inesperado: %q", creado.DepartamentoID)
		}
		if creado.CostoUnitario != 1500 || creado.TipoImpuesto != "IVA" {
			t.Fatalf("campos tributarios inesperados: %+v", creado)
		}
	})

	t.Run("el secuencial es independiente por departamento", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodPost, "/api/productos", cuerpoCreacion(departamentoHogar, ""))
		if rec.Code != http.StatusCreated {
			t.Fatalf("esperaba 201, obtuve %d: %s", rec.Code, rec.Body.String())
		}
		var otro Producto
		if err := json.Unmarshal(rec.Body.Bytes(), &otro); err != nil {
			t.Fatalf("respuesta inválida: %v", err)
		}
		if otro.SKU != "HOG-000002" {
			t.Fatalf("esperaba HOG-000002, obtuve %q", otro.SKU)
		}
	})

	t.Run("obtener", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodGet, "/api/productos/"+creado.ID, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("esperaba 200, obtuve %d", rec.Code)
		}
	})

	t.Run("listar por categoria", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodGet, "/api/productos?categoria_id="+categoriaBebidas, "")
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

	t.Run("listar por departamento", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodGet, "/api/productos?departamento_id="+departamentoHogar, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("esperaba 200, obtuve %d", rec.Code)
		}
		var body ListarProductosBody
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("respuesta inválida: %v", err)
		}
		if body.Total != 2 {
			t.Fatalf("esperaba 2 productos de HOG, obtuve %d", body.Total)
		}
		for _, p := range body.Productos {
			if p.DepartamentoID != departamentoHogar {
				t.Fatalf("producto %s no es del departamento filtrado", p.ID)
			}
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

	t.Run("actualizar no cambia el sku", func(t *testing.T) {
		cuerpo := `{"nombre":"Producto actualizado","costo_unitario":2000,"tipo_impuesto":"EXENTO","permite_venta":false}`
		rec := ejecutar(t, mux, http.MethodPatch, "/api/productos/"+creado.ID, cuerpo)
		if rec.Code != http.StatusOK {
			t.Fatalf("esperaba 200, obtuve %d: %s", rec.Code, rec.Body.String())
		}
		var actualizado Producto
		if err := json.Unmarshal(rec.Body.Bytes(), &actualizado); err != nil {
			t.Fatalf("respuesta inválida: %v", err)
		}
		if actualizado.Nombre != "Producto actualizado" || actualizado.CostoUnitario != 2000 ||
			actualizado.TipoImpuesto != "EXENTO" || actualizado.PermiteVenta {
			t.Fatalf("producto inesperado: %+v", actualizado)
		}
		if actualizado.SKU != creado.SKU {
			t.Fatalf("el sku cambió: %q -> %q", creado.SKU, actualizado.SKU)
		}
	})

	t.Run("el sku no se puede cambiar", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodPatch, "/api/productos/"+creado.ID, `{"sku":"OTRO-1"}`)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("esperaba 422 por campo desconocido, obtuve %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("el departamento no se puede cambiar", func(t *testing.T) {
		rec := ejecutar(t, mux, http.MethodPatch, "/api/productos/"+creado.ID, `{"departamento_id":"`+departamentoHogar+`"}`)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("esperaba 422 por campo desconocido, obtuve %d: %s", rec.Code, rec.Body.String())
		}
	})

	validaciones := []struct {
		nombre string
		cuerpo string
	}{
		{"categorias vacias", `{"departamento_id":"` + departamentoBebidas + `","nombre":"V1","categoria_ids":[],"costo_unitario":1000,"tipo_impuesto":"IVA","permite_venta":true}`},
		{"categorias duplicadas", `{"departamento_id":"` + departamentoBebidas + `","nombre":"V2","categoria_ids":["` + categoriaBebidas + `","` + categoriaBebidas + `"],"costo_unitario":1000,"tipo_impuesto":"IVA","permite_venta":true}`},
		{"categoria uuid invalido", `{"departamento_id":"` + departamentoBebidas + `","nombre":"V3","categoria_ids":["no-es-uuid"],"costo_unitario":1000,"tipo_impuesto":"IVA","permite_venta":true}`},
		{"categoria faltante", `{"departamento_id":"` + departamentoBebidas + `","nombre":"V4","costo_unitario":1000,"tipo_impuesto":"IVA","permite_venta":true}`},
		{"departamento faltante", `{"nombre":"V5","categoria_ids":["` + categoriaBebidas + `"],"costo_unitario":1000,"tipo_impuesto":"IVA","permite_venta":true}`},
		{"departamento uuid invalido", `{"departamento_id":"no-es-uuid","nombre":"V6","categoria_ids":["` + categoriaBebidas + `"],"costo_unitario":1000,"tipo_impuesto":"IVA","permite_venta":true}`},
		{"costo unitario negativo", `{"departamento_id":"` + departamentoBebidas + `","nombre":"V7","categoria_ids":["` + categoriaBebidas + `"],"costo_unitario":-1,"tipo_impuesto":"IVA","permite_venta":true}`},
		{"costo unitario decimal", `{"departamento_id":"` + departamentoBebidas + `","nombre":"V8","categoria_ids":["` + categoriaBebidas + `"],"costo_unitario":1.5,"tipo_impuesto":"IVA","permite_venta":true}`},
		{"tipo de impuesto invalido", `{"departamento_id":"` + departamentoBebidas + `","nombre":"V9","categoria_ids":["` + categoriaBebidas + `"],"costo_unitario":1000,"tipo_impuesto":"OTRO","permite_venta":true}`},
		{"tipo de impuesto faltante", `{"departamento_id":"` + departamentoBebidas + `","nombre":"V10","categoria_ids":["` + categoriaBebidas + `"],"costo_unitario":1000,"permite_venta":true}`},
		{"sku enviado al crear", `{"sku":"HACK-1","departamento_id":"` + departamentoBebidas + `","nombre":"V11","categoria_ids":["` + categoriaBebidas + `"],"costo_unitario":1000,"tipo_impuesto":"IVA","permite_venta":true}`},
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
