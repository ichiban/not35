package handlers

import (
	"context"
	"net/http"

	"github.com/ichiban/not35/app"

	"github.com/gorilla/sessions"
)

type Sessions func(http.Handler) http.Handler

func NewSessions(c *app.Config) Sessions {
	var store = sessions.NewCookieStore([]byte(c.Secret))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "session")
			r = r.WithContext(context.WithValue(r.Context(), sessionKey{}, session))
			next.ServeHTTP(w, r)
		})
	}
}

func session(r *http.Request) *sessions.Session {
	return r.Context().Value(sessionKey{}).(*sessions.Session)
}

type sessionKey struct{}
