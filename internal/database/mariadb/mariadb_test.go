package mariadb

import (
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
)

func TestMariaDB_DatabaseDropQuery(t *testing.T) {
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
			got := ma.DatabaseDropQuery(tt.args.database)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_DumpCommand(t *testing.T) {
	type args struct {
		conf *conftypes.Dump
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"default",
			args{&conftypes.Dump{Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=1.1.1.1", "--user=u", "d", "--verbose"),
		},
		{
			"clean",
			args{&conftypes.Dump{Clean: true, Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=1.1.1.1", "--user=u", "d", "--add-drop-table", "--verbose"),
		},
		{
			"tables",
			args{&conftypes.Dump{Table: []string{"table1", "table2"}, Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=1.1.1.1", "--user=u", "d", "table1", "table2", "--verbose"),
		},
		{
			"exclude-table",
			args{&conftypes.Dump{ExcludeTable: []string{"table1", "table2"}, Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=1.1.1.1", "--user=u", "d", "--ignore-table=table1", "--ignore-table=table2", "--verbose"),
		},
		{
			"port",
			args{&conftypes.Dump{Global: &conftypes.Global{Port: 1234}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb-dump || which mysqldump)"`), "--host=", "--user=", "", "--port=1234", "--verbose"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.DumpCommand(tt.args.conf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_ExecCommand(t *testing.T) {
	type args struct {
		conf *conftypes.Exec
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"default",
			args{&conftypes.Exec{Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "exec", command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d"),
		},
		{
			"disable-headers",
			args{&conftypes.Exec{DisableHeaders: true, Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "exec", command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d", "--skip-column-names"),
		},
		{
			"command",
			args{&conftypes.Exec{Command: "show databases", Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, "exec", command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d", "--execute=show databases"),
		},
		{
			"port",
			args{&conftypes.Exec{Global: &conftypes.Global{Port: 1234}}},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), "exec", command.Raw(`"$(which mariadb || which mysql)"`), "--host=", "--user=", "--port=1234"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.ExecCommand(tt.args.conf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_DatabaseListQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "show databases"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.DatabaseListQuery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_PasswordEnvs(t *testing.T) {
	type args struct {
		c *conftypes.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{"default", args{&conftypes.Global{}}, kubernetes.ConfigLookups{kubernetes.LookupEnv{"MARIADB_PASSWORD", "MYSQL_PASSWORD"}, kubernetes.LookupSecretVolume{Name: "mariadb-credentials", Key: "mariadb-password"}}},
		{"root", args{&conftypes.Global{Username: "root"}}, kubernetes.ConfigLookups{kubernetes.LookupEnv{"MARIADB_ROOT_PASSWORD", "MYSQL_ROOT_PASSWORD"}, kubernetes.LookupSecretVolume{Name: "mariadb-credentials", Key: "mariadb-root-password"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := MariaDB{}
			got := db.PasswordEnvs(tt.args.c)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMariaDB_RestoreCommand(t *testing.T) {
	type args struct {
		conf        *conftypes.Restore
		inputFormat sqlformat.Format
	}
	tests := []struct {
		name string
		args args
		want *command.Builder
	}{
		{
			"gzip",
			args{&conftypes.Restore{Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Gzip},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d"),
		},
		{
			"plain",
			args{&conftypes.Restore{Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Plain},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d"),
		},
		{
			"custom",
			args{&conftypes.Restore{Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u"}}, sqlformat.Custom},
			command.NewBuilder(command.Env{Key: "MYSQL_PWD"}, command.Raw(`"$(which mariadb || which mysql)"`), "--host=1.1.1.1", "--user=u", "--database=d"),
		},
		{
			"port",
			args{&conftypes.Restore{Global: &conftypes.Global{Port: 1234}}, sqlformat.Plain},
			command.NewBuilder(command.NewEnv("MYSQL_PWD", ""), command.Raw(`"$(which mariadb || which mysql)"`), "--host=", "--user=", "--database=", "--port=1234"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MariaDB{}
			got := ma.RestoreCommand(tt.args.conf, tt.args.inputFormat)
			assert.Equal(t, tt.want, got)
		})
	}
}
