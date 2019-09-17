package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/csrf"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-playground/form"
	"github.com/ichiban/assets"
	"github.com/ichiban/not35/models"
	"github.com/ichiban/not35/views"
)

func New(
	l *assets.Locator,
	render func(io.Writer, interface{}),

	csrfProtection CSRFProtection,
	sessions Sessions,
	authentication Authentication,

	userRepo *models.UserRepository,
	noteRepo *models.NoteRepository,
) http.Handler {
	decoder := form.NewDecoder()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(Logger)
	r.Use(middleware.Recoverer)
	r.Use(Override)
	r.Use(csrfProtection)
	r.Use(sessions)

	r.Get(`/sessions/new`, func(w http.ResponseWriter, r *http.Request) {
		s := session(r)

		flashes := s.Flashes()
		if err := s.Save(r, w); err != nil {
			panic(err)
		}

		render(w, &views.SessionNew{
			Flashes: flashes,
			CSRF:    csrf.TemplateField(r),
		})
	})

	r.Post(`/sessions`, func(w http.ResponseWriter, r *http.Request) {
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

		var u models.User
		if err := userRepo.Authenticate(r.Context(), &u, form.Email, form.Password); err != nil {
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
	})

	r.Get(`/sessions/this`, func(w http.ResponseWriter, r *http.Request) {
		render(w, &views.Session{
			CSRF: csrf.TemplateField(r),
		})
	})

	r.Delete(`/sessions/this`, func(w http.ResponseWriter, r *http.Request) {
		s := session(r)
		s.Options.MaxAge = -1
		if err := s.Save(r, w); err != nil {
			panic(err)
		}

		redirectBack(w, r)
	})

	r.Get(`/`, func(w http.ResponseWriter, r *http.Request) {
		var notes []models.Note
		if err := noteRepo.FindAll(r.Context(), &notes); err != nil {
			render(w, err)
			return
		}

		render(w, &notes)
	})

	r.With(authentication).Post(`/`, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			render(w, err)
			return
		}

		var note models.Note
		if err := decoder.Decode(&note, r.Form); err != nil {
			render(w, err)
			return
		}

		if err := noteRepo.Add(r.Context(), &note); err != nil {
			render(w, err)
			return
		}

		render(w, &note)
	})

	r.With(authentication).Get(`/new`, func(w http.ResponseWriter, r *http.Request) {
		render(w, &views.NoteNew{
			CSRF: csrf.TemplateField(r),
		})
	})

	r.Get(`/{ID:\d+}`, func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "ID"))

		var note models.Note
		if err := noteRepo.Find(r.Context(), &note, id); err != nil {
			render(w, err)
			return
		}

		render(w, &note)
	})

	r.With(authentication).Get(`/{ID:\d+}/edit`, func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "ID"))

		note := views.NoteEdit{
			CSRF: csrf.TemplateField(r),
		}
		if err := noteRepo.Find(r.Context(), &note.Note, id); err != nil {
			render(w, err)
			return
		}

		render(w, &note)
	})

	r.With(authentication).Put(`/{ID:\d+}`, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			render(w, err)
			return
		}

		id, _ := strconv.Atoi(chi.URLParam(r, "ID"))

		var note models.Note
		if err := noteRepo.Find(r.Context(), &note, id); err != nil {
			render(w, err)
			return
		}

		if err := decoder.Decode(&note, r.Form); err != nil {
			render(w, err)
			return
		}

		if err := noteRepo.Add(r.Context(), &note); err != nil {
			render(w, err)
			return
		}

		render(w, &note)
	})

	r.Mount("/", http.FileServer(http.Dir(filepath.Join(l.Path, "public"))))

	return r
}

func redirectBack(w http.ResponseWriter, r *http.Request) {
	s := session(r)
	path, _ := s.Values["redirect-to"].(string)
	delete(s.Values, "redirect-to")
	http.Redirect(w, r, path, http.StatusFound)
}
