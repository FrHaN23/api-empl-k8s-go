package modules

import (
	"api-empl-k8s-go/handler"
	repo "api-empl-k8s-go/repos"
	service "api-empl-k8s-go/services"

	"github.com/jmoiron/sqlx"
)

type Module struct {
	Handler *handler.EmployeeHandler
}

func EmployeeModule(db *sqlx.DB) *Module {
	r := repo.NewEmployeeRepo(db)
	s := service.NewEmployeeService(r)
	h := handler.NewEmployeeHandler(s)

	return &Module{Handler: h}
}
