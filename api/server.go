package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grysj/remitly-api/config"
	"github.com/grysj/remitly-api/db"
	"github.com/rs/cors"
)

type Server struct {
	store  *db.Store
	router http.Handler
}

func NewServer(store *db.Store, cfg config.Config) (*Server, error) {

	mux := http.NewServeMux()

	server := &Server{
		store: store,
	}

	mux.HandleFunc("GET /v1/swift-codes/{swiftcode...}", server.getSwiftDetails)
	mux.HandleFunc("GET /v1/swift-codes/country/{countryISO2code...}", server.getSwiftCodes)
	mux.HandleFunc("POST /v1/swift-codes", Middleware(cfg.ApiPassword, server.postSwiftCode))
	mux.HandleFunc("DELETE /v1/swift-codes/{swiftcode...}", Middleware(cfg.ApiPassword, server.deleteSwift))
	mux.HandleFunc("/", server.notFoundHandler)

	c := cors.New(cors.Options{
		AllowedOrigins: cfg.CorsAllowedOrigins,
		AllowedMethods: cfg.CorsAllowedMethods,
		AllowedHeaders: cfg.CorsAllowedHeaders,
	})

	server.router = c.Handler(mux)

	return server, nil

}

func (server *Server) StartServer(port string) error {
	if server.router == nil {
		return fmt.Errorf("server router not initialized")
	}

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting server on %s\n", addr)

	return http.ListenAndServe(addr, server.router)
}

type NotFoundResponse struct {
	Message string `json:"message"`
}

func (server *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Status", fmt.Sprintf("%d", http.StatusNotFound))
	w.WriteHeader(http.StatusNotFound)

	response := NotFoundResponse{
		Message: fmt.Sprintf("Route %s %s not found", r.Method, r.URL.Path),
	}

	json.NewEncoder(w).Encode(response)
}
