package handlers

import (
	"context"
	"net/http"

	"github.com/ichiban/not35/models"
)

type Authentication func(http.Handler) http.Handler

func NewAuthentication(userRepo *models.UserRepository) Authentication {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s := session(r)
			id, _ := s.Values["user-id"].(int)

			var u models.User
			if err := userRepo.Find(r.Context(), &u, id); err != nil {
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
}

type userKey struct{}

func user(r *http.Request) *models.User {
	if u, ok := r.Context().Value(userKey{}).(*models.User); ok {
		return u
	}
	return nil
}
