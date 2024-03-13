package mongodb

import (
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
)

func TestMongoDB_DumpCommand(t *testing.T) {
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
			args{config.Dump{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("mongodump", "--archive", "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--db=d"),
		},
		{
			"clean",
			args{config.Dump{Clean: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("mongodump", "--archive", "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--db=d"),
		},
		{
			"tables",
			args{config.Dump{Tables: []string{"table1"}, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("mongodump", "--archive", "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--db=d", "--collection=table1"),
		},
		{
			"exclude-table",
			args{config.Dump{ExcludeTable: []string{"table1", "table2"}, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("mongodump", "--archive", "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--db=d", "--excludeCollection=table1", "--excludeCollection=table2"),
		},
		{
			"quiet",
			args{config.Dump{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p", Quiet: true}}},
			command.NewBuilder("mongodump", "--archive", "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--db=d", "--quiet"),
		},
		{
			"port",
			args{config.Dump{Global: config.Global{Port: 1234}}},
			command.NewBuilder("mongodump", "--archive", "--host=", "--username=", "--password=", "--authenticationDatabase=", "--port=1234"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			got := ma.DumpCommand(tt.args.conf)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMongoDB_ExecCommand(t *testing.T) {
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
			args{config.Exec{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("exec", command.Raw(`"$(which mongosh || which mongo)"`), "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "d"),
		},
		{
			"disable-headers",
			args{config.Exec{DisableHeaders: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("exec", command.Raw(`"$(which mongosh || which mongo)"`), "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--quiet", "d"),
		},
		{
			"command",
			args{config.Exec{Command: "show databases", Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("exec", command.Raw(`"$(which mongosh || which mongo)"`), "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--eval=show databases", "d"),
		},
		{
			"port",
			args{config.Exec{Global: config.Global{Port: 1234}}},
			command.NewBuilder("exec", command.Raw(`"$(which mongosh || which mongo)"`), "--host=", "--username=", "--password=", "--authenticationDatabase=", "--port=1234"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			got := ma.ExecCommand(tt.args.conf)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMongoDB_PasswordEnvNames(t *testing.T) {
	type args struct {
		c config.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{"default", args{}, kubernetes.ConfigLookups{kubernetes.LookupEnv{"MONGODB_EXTRA_PASSWORDS"}}},
		{"root", args{config.Global{Username: "root"}}, kubernetes.ConfigLookups{kubernetes.LookupEnv{"MONGODB_ROOT_PASSWORD"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := MongoDB{}
			got := db.PasswordEnvNames(tt.args.c)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMongoDB_RestoreCommand(t *testing.T) {
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
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}, sqlformat.Gzip},
			command.NewBuilder("mongorestore", "--archive", "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--db=d"),
		},
		{
			"plain",
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}, sqlformat.Plain},
			command.NewBuilder("mongorestore", "--archive", "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--db=d"),
		},
		{
			"custom",
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}, sqlformat.Custom},
			command.NewBuilder("mongorestore", "--archive", "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--db=d"),
		},
		{
			"clean",
			args{config.Restore{Clean: true, Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"}}, sqlformat.Gzip},
			command.NewBuilder("mongorestore", "--archive", "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--drop", "--db=d"),
		},
		{
			"quiet",
			args{config.Restore{Global: config.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p", Quiet: true}}, sqlformat.Gzip},
			command.NewBuilder("mongorestore", "--archive", "--host=1.1.1.1", "--username=u", "--password=p", "--authenticationDatabase=d", "--db=d", "--quiet"),
		},
		{
			"port",
			args{config.Restore{Global: config.Global{Port: 1234}}, sqlformat.Gzip},
			command.NewBuilder("mongorestore", "--archive", "--host=", "--username=", "--password=", "--authenticationDatabase=", "--port=1234"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			got := ma.RestoreCommand(tt.args.conf, tt.args.inputFormat)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestMongoDB_AuthenticationDatabase(t *testing.T) {
	type args struct {
		c config.Global
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"root", args{config.Global{Host: "1.1.1.1", Username: "root"}}, "admin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := MongoDB{}
			got := db.AuthenticationDatabase(tt.args.c)
			assert.Equal(t, tt.want, got)
		})
	}
}
