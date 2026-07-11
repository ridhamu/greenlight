// Package data
package data

import (
	"database/sql"
	"errors"
)

var (
	ErrNotFoundRecord = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Movies     MovieModel
	Users      UserModel
	TokenModel TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{
			DB: db,
		},
		Users: UserModel{
			DB: db,
		},
		TokenModel: TokenModel{
			DB: db,
		},
	}
}
