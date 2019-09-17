package views

import "html/template"

type SessionNew struct {
	Flashes []interface{}
	CSRF    template.HTML
}

type Session struct {
	CSRF template.HTML
}
