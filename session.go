package main

import (
	"context"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte(secret))

func sessionMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		r = r.WithContext(context.WithValue(r.Context(), sessionKey{}, session))
		next.ServeHTTP(w, r)
	})
}

func session(r *http.Request) *sessions.Session {
	return r.Context().Value(sessionKey{}).(*sessions.Session)
}

type sessionKey struct{}

type NewSession struct {
	Flashes []interface{}
	CSRF    template.HTML
}

type ShowSession struct {
	CSRF template.HTML
}
