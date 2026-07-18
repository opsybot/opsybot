package db

import (
	"embed"
	"io/fs"
)

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

func PostgresMigrations() (fs.FS, error) {
	return fs.Sub(postgresMigrations, "migrations/postgres")
}
