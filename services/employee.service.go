package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"api-empl-k8s-go/models"
	repo "api-empl-k8s-go/repos"
)

var (
	ErrBadRequest = errors.New("bad request")
	ErrNotFound   = errors.New("not found")
)

type EmployeeService struct {
	repo repo.EmployeeRepo
}

func NewEmployeeService(s repo.EmployeeRepo) *EmployeeService {
	return &EmployeeService{repo: s}
}

// Create validates and creates an employee record.
func (s *EmployeeService) Create(ctx context.Context, in *models.Employee) error {
	// Trim and basic validate
	in.Name = strings.TrimSpace(in.Name)
	in.Position = strings.TrimSpace(in.Position)

	if in.Name == "" {
		return fmt.Errorf("%w: name is required", ErrBadRequest)
	}
	if in.Salary < 0 {
		return fmt.Errorf("%w: salary must be >= 0", ErrBadRequest)
	}

	return s.repo.Create(ctx, in)
}

func (s *EmployeeService) Get(ctx context.Context, id int64) (*models.Employee, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repo.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return e, nil
}

func (s *EmployeeService) Update(ctx context.Context, id int64, fields map[string]any) (*models.Employee, error) {
	if len(fields) == 0 {
		return s.Get(ctx, id)
	}

	allowed := map[string]bool{
		"name":     true,
		"position": true,
		"salary":   true,
	}

	// build sanitized map
	updates := make(map[string]any, len(fields))
	for k, v := range fields {
		k = strings.ToLower(strings.TrimSpace(k))
		if !allowed[k] {
			continue
		}
		switch k {
		case "name":
			if str, ok := v.(string); ok {
				str = strings.TrimSpace(str)
				if str == "" {
					return nil, fmt.Errorf("%w: name cannot be empty", ErrBadRequest)
				}
				updates["name"] = str
			} else {
				return nil, fmt.Errorf("%w: invalid type for name", ErrBadRequest)
			}
		case "position":
			if v == nil {
				updates["position"] = ""
				continue
			}
			if str, ok := v.(string); ok {
				updates["position"] = strings.TrimSpace(str)
			} else {
				return nil, fmt.Errorf("%w: invalid type for position", ErrBadRequest)
			}
		case "salary":
			switch vv := v.(type) {
			case float64:
				if vv < 0 {
					return nil, fmt.Errorf("%w: salary must be >= 0", ErrBadRequest)
				}
				updates["salary"] = int(vv)
			case int:
				if vv < 0 {
					return nil, fmt.Errorf("%w: salary must be >= 0", ErrBadRequest)
				}
				updates["salary"] = vv
			case int64:
				if vv < 0 {
					return nil, fmt.Errorf("%w: salary must be >= 0", ErrBadRequest)
				}
				updates["salary"] = int(vv)
			case nil:
				updates["salary"] = 0
			default:
				return nil, fmt.Errorf("%w: invalid type for salary", ErrBadRequest)
			}
		}
	}

	if len(updates) == 0 {
		return s.Get(ctx, id)
	}

	updated, err := s.repo.Update(ctx, id, updates)
	if err != nil {
		if err == repo.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return updated, nil
}

func (s *EmployeeService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if err == repo.ErrNotFound {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *EmployeeService) List(ctx context.Context, limit, offset int) ([]models.Employee, error) {
	return s.repo.List(ctx, limit, offset)
}
