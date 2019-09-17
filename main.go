package main

import (
	"github.com/ichiban/di"
	"github.com/ichiban/not35/app"
	"github.com/ichiban/not35/handlers"
	"github.com/ichiban/not35/models"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	c := di.MustNew(
		app.NewServer,
		app.NewTerminal,
		app.NewConfig,
		app.NewDB,
		app.NewAssets,
		app.NewRender,

		handlers.New,
		handlers.NewCSRFProtection,
		handlers.NewSessions,
		handlers.NewAuthentication,

		models.NewNoteRepository,
		models.NewUserRepository,
	)
	defer c.MustClose()

	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := terminal.Restore(0, oldState); err != nil {
			panic(err)
		}
	}()

	c.MustConsume(func(s *app.Server, t *app.Terminal) {
		go s.Run()
		t.Run()
	})
}
