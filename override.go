package main

import (
	"net/http"
	"strings"
)

func override(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if m := r.FormValue("_method"); m != "" {
				r.Method = strings.ToUpper(m)
			}
		}
		next.ServeHTTP(w, r)
	})
}
