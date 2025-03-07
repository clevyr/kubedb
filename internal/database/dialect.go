package database

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/mariadb"
	"github.com/clevyr/kubedb/internal/database/meilisearch"
	"github.com/clevyr/kubedb/internal/database/mongodb"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/database/redis"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
)

func All() []conftypes.Database {
	return []conftypes.Database{
		postgres.Postgres{},
		mariadb.MariaDB{},
		mongodb.MongoDB{},
		redis.Redis{},
		meilisearch.Meilisearch{},
	}
}

func Names() []string {
	all := All()
	names := make([]string, 0, len(all))
	for _, db := range All() {
		names = append(names, db.Name())
	}
	return names
}

var ErrUnsupportedDatabase = errors.New("unsupported database")

func New(name string) (conftypes.Database, error) {
	for _, db := range All() {
		if name == db.Name() {
			return db, nil
		}
		if dbAlias, ok := db.(conftypes.DBAliaser); ok {
			if slices.Contains(dbAlias.Aliases(), name) {
				return db, nil
			}
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrUnsupportedDatabase, name)
}

func NamesForInterface[T any]() []string {
	all := All()
	names := make([]string, 0, len(all))
	for _, db := range all {
		if _, ok := db.(T); ok {
			names = append(names, db.Name())
		}
	}
	return names
}

func DetectFormat(db conftypes.DBFiler, path string) sqlformat.Format {
	for format, ext := range db.Formats() {
		if strings.HasSuffix(path, ext) {
			return format
		}
	}
	return sqlformat.Unknown
}

func GetExtension(db conftypes.DBFiler, format sqlformat.Format) string {
	if ext, ok := db.Formats()[format]; ok {
		return ext
	}
	return ""
}
