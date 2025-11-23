package handler

import (
	"api-empl-k8s-go/res"
	"api-empl-k8s-go/types"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	data := types.Res{
		Message: "im alive",
	}
	res.ResOkJSON(w, &data)
}
