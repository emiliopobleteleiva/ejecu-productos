// Comando codegen genera contracts/openapi.yaml desde los structs Huma.
//
// Uso, desde apps/api:
//
//	go run ./cmd/codegen
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	humaapi "github.com/danielgtaylor/huma/v2/yaml"
	apiinternal "github.com/emiliopobleteleiva/ejecu-productos/apps/api/internal/api"
)

func main() {
	salida := flag.String("out", "../../contracts/openapi.yaml", "ruta de salida del contrato OpenAPI")
	flag.Parse()

	mux := http.NewServeMux()
	api := apiinternal.NuevaAPI(mux, "dev", "dev")
	apiinternal.Registrar(api, apiinternal.NuevosManejadoresProductos())

	datos, err := json.MarshalIndent(api.OpenAPI(), "", "  ")
	if err != nil {
		log.Fatalf("no se pudo serializar el contrato: %v", err)
	}

	var yamlBuf bytes.Buffer
	if err := humaapi.Convert(&yamlBuf, bytes.NewReader(datos)); err != nil {
		log.Fatalf("no se pudo convertir el contrato a YAML: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(*salida), 0o755); err != nil {
		log.Fatalf("no se pudo crear el directorio de salida: %v", err)
	}
	if err := os.WriteFile(*salida, yamlBuf.Bytes(), 0o644); err != nil {
		log.Fatalf("no se pudo escribir el contrato: %v", err)
	}

	log.Printf("contrato generado en %s", *salida)
}
