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
	Movies      MovieModel
	Users       UserModel
	Token       TokenModel
	Permissions PermissionsModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{
			DB: db,
		},
		Users: UserModel{
			DB: db,
		},
		Token: TokenModel{
			DB: db,
		},
		Permissions: PermissionsModel{
			DB: db,
		},
	}
}
