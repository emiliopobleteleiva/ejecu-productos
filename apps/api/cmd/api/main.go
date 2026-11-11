package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiinternal "github.com/emiliopobleteleiva/ejecu-productos/apps/api/internal/api"
)

var version = "dev"

func main() {
	puerto := os.Getenv("API_PORT")
	if puerto == "" {
		puerto = "8080"
	}
	entorno := os.Getenv("API_ENV")
	if entorno == "" {
		entorno = "dev"
	}

	mux := http.NewServeMux()
	api := apiinternal.NuevaAPI(mux, version, entorno)
	apiinternal.Registrar(api, apiinternal.NuevosManejadoresProductos())

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		// TODO Hito 2: responder 200 solo si PostgreSQL contesta.
		w.WriteHeader(http.StatusOK)
	})

	servidor := &http.Server{
		Addr:              ":" + puerto,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()

	go func() {
		log.Printf("servicio de productos escuchando en :%s (entorno=%s, versión=%s)", puerto, entorno, version)
		if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error al iniciar el servidor: %v", err)
		}
	}()

	<-ctx.Done()
	ctxApagado, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	if err := servidor.Shutdown(ctxApagado); err != nil {
		log.Printf("error al apagar el servidor: %v", err)
	}
}
