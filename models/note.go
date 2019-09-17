package models

import (
	"context"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"golang.org/x/net/html"
)

type Note struct {
	ID        int       `db:"id"`
	Body      string    `db:"body"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (n *Note) Snippet(l int) (string, error) {
	// wrapped by <html><head></head><body>...</body></html>
	d, err := html.Parse(strings.NewReader(n.Body))
	if err != nil {
		return "", err
	}

	_ = truncateHTML(d, l)

	// strip <html><head></head><body>...</body></html> and render
	var w strings.Builder
	for n := d.LastChild.LastChild.FirstChild; n != nil; n = n.NextSibling {
		if err := html.Render(&w, n); err != nil {
			return "", err
		}
	}

	return w.String(), nil
}

func truncateHTML(n *html.Node, l int) int {
	if l <= 0 {
		n.Parent.RemoveChild(n)
		return l
	}

	if n.Type == html.TextNode {
		r := []rune(n.Data)
		if l < len(r) {
			r = r[:l]
			r[l-1] = '…'
			n.Data = string(r)
			return 0
		}

		return l - len(n.Data)
	}

	var cs []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		cs = append(cs, c)
	}

	for _, c := range cs {
		l = truncateHTML(c, l)
	}

	return l
}

type NoteRepository struct {
	db *sqlx.DB
}

func NewNoteRepository(db *sqlx.DB) *NoteRepository {
	return &NoteRepository{
		db: db,
	}
}

func (r *NoteRepository) Find(ctx context.Context, n *Note, id int) error {
	return r.db.GetContext(ctx, n, "SELECT * FROM notes WHERE id = $1", id)
}

func (r *NoteRepository) FindAll(ctx context.Context, ns *[]Note) error {
	return r.db.SelectContext(ctx, ns, "SELECT * FROM notes ORDER BY id DESC")
}

func (r *NoteRepository) Add(ctx context.Context, n *Note) error {
	if n.ID == 0 {
		return r.db.GetContext(ctx, n, "INSERT INTO notes (body) VALUES ($1) RETURNING *", n.Body)
	}
	return r.db.GetContext(ctx, n, "UPDATE notes SET body = $1, updated_at = current_timestamp WHERE id = $2 RETURNING *", n.Body, n.ID)
}
