package models

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jmoiron/sqlx"
)

type User struct {
	ID           int       `db:"id"`
	Email        string    `db:"email"`
	PasswordHash []byte    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Add(ctx context.Context, u *User, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return r.db.GetContext(ctx, u, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING *", u.Email, hash)
}

func (r *UserRepository) Find(ctx context.Context, u *User, id int) error {
	return r.db.GetContext(ctx, u, "SELECT * FROM users WHERE id = $1", id)
}

func (r *UserRepository) Authenticate(ctx context.Context, u *User, email, password string) error {
	var c User
	if err := r.db.GetContext(ctx, &c, "SELECT * FROM users WHERE email = $1", email); err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(c.PasswordHash, []byte(password)); err != nil {
		return err
	}

	*u = c

	return nil
}
