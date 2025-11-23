package models

import "time"

type Employee struct {
	ID       int64  `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Position string `db:"position" json:"position"`
	Salary   int    `db:"salary" json:"salary"`

	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
