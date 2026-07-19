package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/ridhamu/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var AnonymousUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

type UserModel struct {
	DB *sql.DB
}

var ErrDuplicateEmail = errors.New("duplicate email")

func (u *User) IsAnonymousUser() bool {
	return u == AnonymousUser
}

func (p *password) Set(plainTextPassword string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plainTextPassword
	p.hash = hashed

	return nil
}

func (p *password) Matches(plainTextPassowrd string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPassowrd))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlain(v *validator.Validator, passwordPlain string) {
	v.Check(passwordPlain != "", "password", "must be provided")
	v.Check(len(passwordPlain) >= 8, "password", "must be more than 8 bytes long")
	v.Check(len(passwordPlain) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	// validate email
	ValidateEmail(v, user.Email)

	// validate password
	if user.Password.plaintext != nil {
		ValidatePasswordPlain(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func (u UserModel) Insert(user *User) error {
	stmt := `INSERT INTO users (name, email, password_hash, activated) VALUES ($1, $2, $3, $4) RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	err := u.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.CreatedAt, &user.version)
	if err != nil {
		var pqError *pq.Error
		switch {
		case errors.As(err, &pqError) && pqError.Code == "23505":
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (u UserModel) GetByEmail(email string) (*User, error) {
	stmt := `SELECT id, created_at, name, email, password_hash, activated, version FROM users WHERE email = $1`

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	var user User

	err := u.DB.QueryRowContext(ctx, stmt, email).Scan(&user.ID, &user.CreatedAt, &user.Name, &user.Email, &user.Password.hash, &user.Activated, &user.version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFoundRecord
		default:
			return nil, err

		}
	}

	return &user, nil
}

func (u UserModel) Update(user *User) error {
	stmt := `UPDATE users SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1 WHERE id = $5 AND version = $6 RETURNING version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated, user.ID, user.version}

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	err := u.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.version)
	if err != nil {
		var pqError *pq.Error
		switch {
		case errors.As(err, &pqError) && pqError.Code == "23505":
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (u UserModel) GetForToken(scope, plainTextToken string) (*User, error) {
	tokenArr := sha256.Sum256([]byte(plainTextToken))

	stmt := `SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version FROM users INNER JOIN tokens ON users.id = tokens.user_id WHERE tokens.hash = $1 AND tokens.scope = $2 AND tokens.expiry > $3`

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	tokenSlice := tokenArr[:]
	args := []any{tokenSlice, scope, time.Now()}

	var user User

	err := u.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.CreatedAt, &user.Name, &user.Email, &user.Password.hash, &user.Activated, &user.version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFoundRecord
		default:
			return nil, err
		}
	}

	return &user, nil
}
