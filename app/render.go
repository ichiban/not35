package app

import (
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/ichiban/render"

	"github.com/ichiban/assets"
)

func NewRender(l *assets.Locator) func(io.Writer, interface{}) {
	m := manifest(l.Path)
	return render.New(path.Join(l.Path, "templates"), template.FuncMap{
		"raw": func(s string) template.HTML {
			return template.HTML(s)
		},
		"rfc3339": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"datetime": func(t time.Time) string {
			return t.Format("2006/01/02 15:04")
		},
		"actual": func(s string) string {
			return m[s]
		},
	})
}

func manifest(path string) map[string]string {
	f, err := os.Open(filepath.Join(path, "public", "manifest.json"))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		panic(err)
	}

	return m
}
