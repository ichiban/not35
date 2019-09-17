package views

import (
	"html/template"

	"github.com/ichiban/not35/models"
)

type NoteNew struct {
	models.Note
	CSRF template.HTML
}

type NoteEdit struct {
	models.Note
	CSRF template.HTML
}
