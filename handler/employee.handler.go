package handler

import (
	"errors"
	"net/http"
	"strconv"

	global "api-empl-k8s-go/const"
	"api-empl-k8s-go/models"
	repo "api-empl-k8s-go/repos"
	"api-empl-k8s-go/res"
	service "api-empl-k8s-go/services"
	"api-empl-k8s-go/types"
	"api-empl-k8s-go/utils"

	"github.com/gorilla/mux"
)

type EmployeeHandler struct {
	services *service.EmployeeService
}

func NewEmployeeHandler(s *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{services: s}
}

// Create employee
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, global.MaxBodyBytes)
	defer r.Body.Close()

	var in models.Employee
	if err := utils.DecodeJSON(&in, r.Body, false); err != nil {
		if he, ok := err.(utils.HTTPError); ok {
			res.ResErrJson(w, he.Status(), errors.New(he.Error()))
			return
		}
		res.ResErrJson(w, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	ctx := r.Context()
	if err := h.services.Create(ctx, &in); err != nil {
		res.ResErrJson(w, http.StatusBadRequest, err)
		return
	}

	res.ResOkJSON(w, &types.Res{
		Message: "ok",
		Data:    in,
	})
}

// Get by id: /employees/{id}
func (h *EmployeeHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok || idStr == "" {
		res.ResErrJson(w, http.StatusBadRequest, errors.New("missing id"))
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		res.ResErrJson(w, http.StatusBadRequest, err)
		return
	}
	ctx := r.Context()
	e, err := h.services.Get(ctx, id)
	if err != nil {
		if err == service.ErrNotFound || err == repo.ErrNotFound {
			res.ResErrJson(w, http.StatusNotFound, errors.New("employee not found"))
			return
		}
		res.ResErrJson(w, http.StatusInternalServerError, err)
		return
	}
	res.ResOkJSON(w, &types.Res{
		Message: "ok",
		Data:    e,
	})
}

// Update: expects JSON with fields to update
func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok || idStr == "" {
		res.ResErrJson(w, http.StatusBadRequest, errors.New("missing id"))
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		res.ResErrJson(w, http.StatusBadRequest, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, global.MaxBodyBytes)
	defer r.Body.Close()

	var fields map[string]any
	if err := utils.DecodeJSON(&fields, r.Body, false); err != nil {
		if he, ok := err.(utils.HTTPError); ok {
			res.ResErrJson(w, he.Status(), errors.New(he.Error()))
			return
		}
		res.ResErrJson(w, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	ctx := r.Context()
	updated, err := h.services.Update(ctx, id, fields)
	if err != nil {
		if err == repo.ErrNotFound {
			res.ResErrJson(w, http.StatusNotFound, err)
			return
		}
		res.ResErrJson(w, http.StatusInternalServerError, err)
		return
	}
	res.ResOkJSON(w, &types.Res{
		Message: "ok",
		Data:    updated,
	})
}

// Delete by id
func (h *EmployeeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok || idStr == "" {
		res.ResErrJson(w, http.StatusBadRequest, errors.New("missing id"))
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		res.ResErrJson(w, http.StatusBadRequest, err)
		return
	}
	if err := h.services.Delete(r.Context(), id); err != nil {
		if err == repo.ErrNotFound {
			res.ResErrJson(w, http.StatusNotFound, err)
			return
		}
		res.ResErrJson(w, http.StatusInternalServerError, err)
		return
	}
	res.ResOkJSON(w, &types.Res{
		Message: "ok",
	})
}

// List employees: /employees?limit=20&offset=0
func (h *EmployeeHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	list, err := h.services.List(r.Context(), limit, offset)
	if err != nil {
		res.ResErrJson(w, http.StatusInternalServerError, err)
		return
	}
	res.ResOkJSON(w, &types.Res{
		Message: "ok",
		Data:    list,
	})
}
