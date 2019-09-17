package app

import (
	"flag"
	"os"
)

type Config struct {
	Database string
	Secret   string
	Host     string
	Bind     string
}

func NewConfig() *Config {
	var c Config

	flag.StringVar(&c.Database, "database", os.Getenv("DATABASE"), ``)
	flag.StringVar(&c.Secret, "secret", os.Getenv("SECRET"), ``)
	flag.StringVar(&c.Host, "host", os.Getenv("HOST"), ``)
	flag.StringVar(&c.Bind, "bind", os.Getenv("BIND"), ``)
	flag.Parse()

	return &c
}
