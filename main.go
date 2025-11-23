package main

import (
	"api-empl-k8s-go/db"
	"api-empl-k8s-go/handler"
	"api-empl-k8s-go/modules"
	"api-empl-k8s-go/routes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	router, cleanup, err := buildApp()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	router.Use(mux.CORSMethodMiddleware(router)) //preflight

	srv := &http.Server{
		Addr:         ":5000",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Println("listening:", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func buildApp() (*mux.Router, func() error, error) {
	db, err := db.Conn()
	if err != nil {
		return nil, nil, err
	}

	empMod := modules.EmployeeModule(db)

	// routing
	router := mux.NewRouter()
	routing := &routes.Routing{
		HealthFunc: handler.HealthCheck,
		Employee:   empMod.Handler,
	}
	routes.NewRoutes(router, routing)

	cleanup := func() error {
		return db.Close()
	}
	return router, cleanup, nil
}
