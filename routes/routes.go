package routes

import (
	"api-empl-k8s-go/handler"
	"net/http"

	"github.com/gorilla/mux"
)

type Routing struct {
	HealthFunc func(http.ResponseWriter, *http.Request)
	Employee   *handler.EmployeeHandler
}

func NewRoutes(router *mux.Router, r *Routing) {
	if router == nil {
		panic("router is nil")
	}
	if r == nil {
		r = &Routing{}
	}

	// Health
	if r.HealthFunc != nil {
		router.HandleFunc("/health", r.HealthFunc).Methods("GET")
	} else {
		router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	}

	EmployeeRoutes(router, r.Employee)
}
