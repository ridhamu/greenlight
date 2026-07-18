package main

import (
	"context"
	"net/http"

	"github.com/ridhamu/greenlight/internal/data"
)

type ContextKey string

const UserContextKey = ContextKey("user")

func (app *application) contextSetUser(r *http.Request, userData *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), UserContextKey, userData)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	currentUser, ok := r.Context().Value(UserContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return currentUser
}
