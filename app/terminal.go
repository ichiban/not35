package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/ichiban/not35/models"

	"github.com/ichiban/linesqueak"
)

type Terminal struct {
	linesqueak.Editor

	userRepo *models.UserRepository
}

func NewTerminal(userRepo *models.UserRepository) *Terminal {
	return &Terminal{
		Editor: linesqueak.Editor{
			In:     bufio.NewReader(os.Stdin),
			Out:    bufio.NewWriter(os.Stdout),
			Prompt: "not35> ",
		},
		userRepo: userRepo,
	}
}

func (t *Terminal) Run() {
	for {
		line, err := t.Line()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("failed to read line")
			return
		}

		if _, err := fmt.Fprintf(t, "%s%s\n", t.Prompt, line); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("failed to print")
		}

		cmd := strings.Fields(line)
		if len(cmd) == 0 {
			continue
		}

		switch cmd[0] {
		case "add-user":
			if len(cmd) != 3 {
				if _, err := fmt.Fprintf(t, "add-user <email> <password>\n"); err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("failed to print")
				}
				continue
			}
			u := models.User{
				Email: cmd[1],
			}
			if err := t.userRepo.Add(context.Background(), &u, cmd[2]); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("failed to add user")
			}
			if _, err := fmt.Fprintf(t, "created: %s\n", u.Email); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("failed to print")
			}
		default:
			if _, err := fmt.Fprintf(t, "command not found\n"); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("failed to print")
			}
		}
	}
}
