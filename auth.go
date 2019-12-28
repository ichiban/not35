package main

import (
	"context"
	"net/http"
)

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := session(r)
		id, _ := s.Values["user-id"].(int)

		var u User
		if err := db.GetContext(r.Context(), &u, "SELECT * FROM users WHERE id = $1", id); err != nil {
			s.Values["redirect-to"] = r.URL.String()
			if err := s.Save(r, w); err != nil {
				panic(err)
			}
			http.Redirect(w, r, "/sessions/new", http.StatusTemporaryRedirect)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), userKey{}, &u))
		next.ServeHTTP(w, r)
	})
}

type userKey struct{}

func user(r *http.Request) *User {
	if u, ok := r.Context().Value(userKey{}).(*User); ok {
		return u
	}
	return nil
}
