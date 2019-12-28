package main

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int       `db:"id"`
	Email        string    `db:"email"`
	PasswordHash []byte    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func authenticate(ctx context.Context, u *User, email, password string) error {
	var c User
	if err := db.GetContext(ctx, &c, "SELECT * FROM users WHERE email = $1", email); err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(c.PasswordHash, []byte(password)); err != nil {
		return err
	}

	*u = c

	return nil
}
