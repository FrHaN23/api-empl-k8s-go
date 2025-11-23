package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"api-empl-k8s-go/models"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("not found")

type EmployeeRepo interface {
	Create(ctx context.Context, e *models.Employee) error
	GetByID(ctx context.Context, id int64) (*models.Employee, error)
	Update(ctx context.Context, id int64, fields map[string]any) (*models.Employee, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]models.Employee, error)
}

type SQLXEmployeeRepo struct {
	db *sqlx.DB
}

func NewEmployeeRepo(db *sqlx.DB) *SQLXEmployeeRepo { return &SQLXEmployeeRepo{db: db} }

func (r *SQLXEmployeeRepo) Create(ctx context.Context, e *models.Employee) error {
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now

	query := `
INSERT INTO employees (name, position, salary, created_at, updated_at)
VALUES (:name, :position, :salary, :created_at, :updated_at)
RETURNING id
`
	rows, err := r.db.NamedQueryContext(ctx, query, e)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&e.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *SQLXEmployeeRepo) GetByID(ctx context.Context, id int64) (*models.Employee, error) {
	var e models.Employee
	if err := r.db.GetContext(ctx, &e, "SELECT * FROM employees WHERE id=$1 AND deleted_at IS NULL", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *SQLXEmployeeRepo) Update(ctx context.Context, id int64, fields map[string]any) (*models.Employee, error) {
	if len(fields) == 0 {
		return r.GetByID(ctx, id)
	}
	// ensure updated_at is always set
	fields["updated_at"] = time.Now().UTC()

	// build SET clause and args with correct parameter positions
	setParts := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+1)
	i := 1
	for k, v := range fields {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}
	// id will be the last parameter
	args = append(args, id)
	query := "UPDATE employees SET " + strings.Join(setParts, ", ") + fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL RETURNING *", i)

	var updated models.Employee
	if err := r.db.GetContext(ctx, &updated, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &updated, nil
}

func (r *SQLXEmployeeRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "UPDATE employees SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SQLXEmployeeRepo) List(ctx context.Context, limit, offset int) ([]models.Employee, error) {
	var list []models.Employee
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM employees WHERE deleted_at IS NULL ORDER BY id DESC LIMIT $1 OFFSET $2", limit, offset); err != nil {
		return nil, err
	}
	return list, nil
}
