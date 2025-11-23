package routes

import (
	"api-empl-k8s-go/handler"

	"github.com/gorilla/mux"
)

func EmployeeRoutes(mux *mux.Router, h *handler.EmployeeHandler) {
	prefix := mux.PathPrefix("/employees").Subrouter()
	// prefix.Use(middleware.Middleware)
	{
		prefix.HandleFunc("", h.List).Methods("GET")
		prefix.HandleFunc("", h.Create).Methods("POST")
		prefix.HandleFunc("/{id}", h.Get).Methods("GET")
		prefix.HandleFunc("/{id}", h.Update).Methods("PUT")
		prefix.HandleFunc("/{id}", h.Delete).Methods("DELETE")
	}
}
