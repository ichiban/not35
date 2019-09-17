package handlers

import (
	"net/http"

	"github.com/ichiban/not35/app"

	"github.com/gorilla/csrf"
)

type CSRFProtection func(http.Handler) http.Handler

func NewCSRFProtection(c *app.Config) CSRFProtection {
	return csrf.Protect([]byte(c.Secret), csrf.Secure(c.Host != ""))
}
