package main

import (
	"context"
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-playground/form"
	"github.com/gorilla/csrf"
	"github.com/ichiban/assets"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

var (
	database string
	secret   string
	host     string
	bind     string
)

var db *sqlx.DB

var locator *assets.Locator

func init() {
	var err error
	locator, err = assets.New()
	if err != nil {
		panic(err)
	}
}

var decoder = form.NewDecoder()

func main() {
	flag.StringVar(&database, "database", os.Getenv("DATABASE"), ``)
	flag.StringVar(&secret, "secret", os.Getenv("SECRET"), ``)
	flag.StringVar(&host, "host", os.Getenv("HOST"), ``)
	flag.StringVar(&bind, "bind", os.Getenv("BIND"), ``)
	flag.Parse()

	db = sqlx.MustOpen("postgres", database)
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logger)
	r.Use(middleware.Recoverer)
	r.Use(override)
	r.Use(csrf.Protect([]byte(secret), csrf.Secure(host != "")))
	r.Use(sessionMiddle)

	r.Get(`/sessions/new`, newSession)
	r.Post(`/sessions`, createSession)
	r.Get(`/sessions/this`, showSession)
	r.Delete(`/sessions/this`, deleteSession)

	r.Get(`/`, listNotes)
	r.With(auth).Post(`/`, createNote)
	r.With(auth).Get(`/new`, newNote)
	r.Get(`/{ID:\d+}`, showNote)
	r.With(auth).Get(`/{ID:\d+}/edit`, editNote)
	r.With(auth).Put(`/{ID:\d+}`, updateNote)

	r.Mount("/", http.FileServer(http.Dir(filepath.Join(locator.Path, "public"))))

	s := http.Server{
		Addr:      bind,
		TLSConfig: tlsConfig(),
		Handler:   r,
	}

	go func() {
		if s.TLSConfig == nil {
			if err := s.ListenAndServe(); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("stopped listen and serve")
			}
		} else {
			if err := s.ListenAndServeTLS("", ""); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("stopped listen and serve")
			}
		}
	}()

	<-q

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to shutdown")
	}
}

func tlsConfig() *tls.Config {
	if host == "" {
		return nil
	}

	m := autocert.Manager{
		Cache:      autocert.DirCache(path.Join(locator.Path, "certs")),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(host),
	}

	return m.TLSConfig()
}

func redirectBack(w http.ResponseWriter, r *http.Request) {
	s := session(r)
	path, _ := s.Values["redirect-to"].(string)
	delete(s.Values, "redirect-to")
	http.Redirect(w, r, path, http.StatusFound)
}

func newSession(w http.ResponseWriter, r *http.Request) {
	s := session(r)

	flashes := s.Flashes()
	if err := s.Save(r, w); err != nil {
		panic(err)
	}

	render(w, &NewSession{
		Flashes: flashes,
		CSRF:    csrf.TemplateField(r),
	})
}

func createSession(w http.ResponseWriter, r *http.Request) {
	time.Sleep(3 * time.Second)

	if err := r.ParseForm(); err != nil {
		render(w, err)
		return
	}

	var form struct {
		Email    string
		Password string
	}
	if err := decoder.Decode(&form, r.Form); err != nil {
		render(w, err)
		return
	}

	s := session(r)

	var u User
	if err := authenticate(r.Context(), &u, form.Email, form.Password); err != nil {
		s.AddFlash("failed to login")
		if err := s.Save(r, w); err != nil {
			panic(err)
		}
		http.Redirect(w, r, "/sessions/new", http.StatusFound)
		return
	}

	s.Values["user-id"] = u.ID
	if err := s.Save(r, w); err != nil {
		panic(err)
	}

	redirectBack(w, r)
}

func showSession(w http.ResponseWriter, r *http.Request) {
	render(w, &ShowSession{
		CSRF: csrf.TemplateField(r),
	})
}

func deleteSession(w http.ResponseWriter, r *http.Request) {
	s := session(r)
	s.Options.MaxAge = -1
	if err := s.Save(r, w); err != nil {
		panic(err)
	}

	redirectBack(w, r)
}

func listNotes(w http.ResponseWriter, r *http.Request) {
	var notes []Note
	if err := db.SelectContext(r.Context(), &notes, "SELECT * FROM notes ORDER BY id DESC"); err != nil {
		render(w, err)
		return
	}

	render(w, notes)
}

func createNote(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		render(w, err)
		return
	}

	var note Note
	if err := decoder.Decode(&note, r.Form); err != nil {
		render(w, err)
		return
	}

	if err := db.GetContext(r.Context(), &note, "INSERT INTO notes (body) VALUES ($1) RETURNING *", note.Body); err != nil {
		render(w, err)
		return
	}

	render(w, &note)
}

func newNote(w http.ResponseWriter, r *http.Request) {
	render(w, &NewNote{
		CSRF: csrf.TemplateField(r),
	})
}

func showNote(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "ID"))

	var note Note
	if err := db.GetContext(r.Context(), &note, "SELECT * FROM notes WHERE id = $1", id); err != nil {
		render(w, err)
		return
	}

	render(w, &note)
}

func editNote(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "ID"))

	note := EditNote{
		CSRF: csrf.TemplateField(r),
	}
	if err := db.GetContext(r.Context(), &note, "SELECT * FROM notes WHERE id = $1", id); err != nil {
		render(w, err)
		return
	}

	render(w, &note)
}

func updateNote(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		render(w, err)
		return
	}

	id, _ := strconv.Atoi(chi.URLParam(r, "ID"))

	var note Note
	if err := db.GetContext(r.Context(), &note, "SELECT * FROM notes WHERE id = $1", id); err != nil {
		render(w, err)
		return
	}

	if err := decoder.Decode(&note, r.Form); err != nil {
		render(w, err)
		return
	}

	if err := db.GetContext(r.Context(), &note, "UPDATE notes SET body = $1, updated_at = current_timestamp WHERE id = $2 RETURNING *", note.Body, note.ID); err != nil {
		render(w, err)
		return
	}

	render(w, &note)
}
