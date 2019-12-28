package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

var templates *template.Template

func init() {
	m := manifest()

	templates = template.Must(template.New("").Funcs(template.FuncMap{
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
	}).ParseGlob(filepath.Join(locator.Path, "templates", "*.tmpl")))
}

func render(w io.Writer, c interface{}) {
	ts := template.Must(templates.Clone())

	var name string
	if t, ok := c.(Templater); ok {
		name = t.Template()
	} else {
		name = templateName(c)
	}

	layout := "layout.tmpl"
	if l, ok := c.(Layouter); ok {
		layout = l.Layout()
	}

	t := ts.Lookup(name)
	if t == nil {
		panic(fmt.Errorf("template not found: %s", name))
	}

	t = template.Must(t.AddParseTree("content", t.Tree))
	if err := t.ExecuteTemplate(w, layout, c); err != nil {
		panic(err)
	}
}

func manifest() map[string]string {
	f, err := os.Open(filepath.Join(locator.Path, "public", "manifest.json"))
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

func templateName(c interface{}) string {
	t := reflect.TypeOf(c)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return fmt.Sprintf("%s.tmpl", t)
}

// Templater is an interface of view objects which have a custom template name other than <type>.tmpl.
type Templater interface {
	Template() string
}

// Layouter is an interface of view objects which have a custom layout other than layout.tmpl.
type Layouter interface {
	Layout() string
}
