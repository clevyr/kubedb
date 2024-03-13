package mariadb

import (
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
)

func TestMariaDB_DropDatabaseQuery(t *testing.T) {
	type args struct {
		database string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"database", args{"database"}, "set FOREIGN_KEY_CHECKS=0; create or replace database `database`; set FOREIGN_KEY_CHECKS=1; use `database`;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.DropDatabaseQuery(tt.args.database)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_DumpCommand(t *testing.T) {
	type args struct {
		conf config.Dump
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"default",
			args{config.Dump{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=1.1.1.1", "--user=u", "d", "--verbose"),
		},
		{
			"clean",
			args{config.Dump{Clean: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=1.1.1.1", "--user=u", "d", "--add-drop-table", "--verbose"),
		},
		{
			"tables",
			args{config.Dump{Tables: []string{"table1", "table2"}, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=1.1.1.1", "--user=u", "d", "table1", "table2", "--verbose"),
		},
		{
			"exclude-table",
			args{config.Dump{ExcludeTable: []string{"table1", "table2"}, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=1.1.1.1", "--user=u", "d", "--ignore-table=table1", "--ignore-table=table2", "--verbose"),
		},
		{
			"port",
			args{config.Dump{Global: config.Global{Port: 1234}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=", "--user=", "", "--port=1234", "--verbose"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.DumpCommand(tt.args.conf)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMariaDB_ExecCommand(t *testing.T) {
	type args struct {
		conf config.Exec
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"default",
			args{config.Exec{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "exec", command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d"),
		},
		{
			"disable-headers",
			args{config.Exec{DisableHeaders: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "exec", command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d", "--skip-column-names"),
		},
		{
			"command",
			args{config.Exec{Command: "show databases", Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "exec", command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d", "--execute=show databases"),
		},
		{
			"port",
			args{config.Exec{Global: config.Global{Port: 1234}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), "exec", command.Raw(`"$(which mariadb || which mysql)"`), "--host=", "--user=", "--port=1234"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.ExecCommand(tt.args.conf)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMariaDB_ListDatabasesQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "show databases"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.ListDatabasesQuery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_PasswordEnvNames(t *testing.T) {
	type args struct {
		c config.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{"default", args{}, kubernetes.ConfigLookups{kubernetes.LookupEnv{"MARIADB_PASSWORD", "MYSQL_PASSWORD"}}},
		{"root", args{config.Global{Username: "root"}}, kubernetes.ConfigLookups{kubernetes.LookupEnv{"MARIADB_ROOT_PASSWORD", "MYSQL_ROOT_PASSWORD"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := MariaDB{}
			got := db.PasswordEnvNames(tt.args.c)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_RestoreCommand(t *testing.T) {
	type args struct {
		conf        config.Restore
		inputFormat sqlformat.Format
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"gzip",
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Gzip},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d"),
		},
		{
			"plain",
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Plain},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d"),
		},
		{
			"custom",
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d"),
		},
		{
			"port",
			args{config.Restore{Global: config.Global{Port: 1234}}, sqlformat.Plain},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb || which mysql)"`), "--host=", "--user=", "--database=", "--port=1234"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.RestoreCommand(tt.args.conf, tt.args.inputFormat)
			assert.Equal(t, got, tt.want)
		})
	}
}
