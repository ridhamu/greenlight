package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/ridhamu/greenlight/internal/validator"
)

const (
	ScopeActivation = "activation"
)

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scrope    string
}

type TokenModel struct {
	DB *sql.DB
}

func ValidateTokenPlainText(v *validator.Validator, tokenPlainText string) {
	v.Check(tokenPlainText != "", "token", "field must be filled")
	v.Check(len(tokenPlainText) == 26, "token", "token must be 26 byte long")
}

func (t TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	generatedToken, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	// perform insertion here
	err = t.Insert(generatedToken)

	return generatedToken, err
}

func (t TokenModel) Insert(token *Token) error {
	stmt := `INSERT INTO tokens (hash, user_id, expiry, scope) VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scrope}

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	_, err := t.DB.ExecContext(ctx, stmt, args...)
	if err != nil {
		return err
	}
	return nil
}

func (t TokenModel) DeleteAllForUser(userID int64, scope string) error {
	stmt := `DELETE FROM tokens WHERE scope = $1 AND user_id = $2`

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	_, err := t.DB.ExecContext(ctx, stmt, scope, userID)
	return err
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scrope: scope,
	}

	// initialized a zero-valued byte slice with a light of 16 bytes
	randomBytes := make([]byte, 16)

	// fill that randombytes with the help of our OS
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// using that randombytes that we just filled with the help of the OS
	// we want to generate the token.PlainText
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// now we have the plaintext, we could generate the hash value of it
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}
