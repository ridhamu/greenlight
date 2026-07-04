// Package data
package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/ridhamu/greenlight/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

type MovieModel struct {
	DB *sql.DB
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year > 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime >= 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

func (m MovieModel) Insert(movie *Movie) error {
	stmt := `INSERT INTO movies (title, year, runtime, genres) VALUES ($1, $2, $3, $4) RETURNING id, created_at, version `

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrNotFoundRecord
	}

	stmt := `SELECT id, created_at, title, year, runtime, genres, version FROM movies WHERE id = $1`

	var movie Movie

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	err := m.DB.QueryRowContext(ctx, stmt, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFoundRecord
		default:
			return nil, err
		}
	}
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	stmt := `UPDATE movies SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1 WHERE id = $5 AND version = $6 RETURNING version`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFoundRecord):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrNotFoundRecord
	}

	stmt := ` DELETE FROM movies WHERE id = $1 `

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	r, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	affectedRow, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if affectedRow == 0 {
		return ErrNotFoundRecord
	}

	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	stmt := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		ORDER BY id
	`

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	r, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	movieList := []*Movie{}

	for r.Next() {
		var movie Movie

		err = r.Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
		if err != nil {
			return nil, err
		}
		movieList = append(movieList, &movie)
	}

	if err = r.Err(); err != nil {
		return nil, err
	}

	return movieList, nil
}
