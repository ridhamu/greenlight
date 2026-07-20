package data

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

type Permissions []string

func (p Permissions) Include(code string) bool { // why done
	return slices.Contains(p, code)
}

type PermissionsModel struct {
	DB *sql.DB
}

func (p PermissionsModel) AddForUser(userID int64, codes ...string) error {
	stmt := `
			INSERT INTO
				users_permissions
			SELECT
				$1,
				permissions.id
			FROM
				permissions
			WHERE
				permissions.code = ANY($2)
	`

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	_, err := p.DB.ExecContext(ctx, stmt, userID, pq.Array(codes))
	return err
}

func (p PermissionsModel) GetAllForUser(userID int64) (Permissions, error) {
	stmt := `
			SELECT
				permissions.code
			FROM
				permissions
				INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
				INNER JOIN users ON users_permissions.user_id = users.id
			WHERE
				users.id = $1`

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Hour)
	defer cf()

	rows, err := p.DB.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}

	var permissions Permissions

	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
