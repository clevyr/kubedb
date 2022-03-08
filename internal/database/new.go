package database

import (
	"errors"
	"fmt"
)

var ErrUnsupportedDatabase = errors.New("unsupported database")

func New(name string) (Databaser, error) {
	switch name {
	case "postgresql", "postgres", "psql", "pg":
		return Postgres{}, nil
	}
	return nil, fmt.Errorf("%v: %s", ErrUnsupportedDatabase, name)
}
