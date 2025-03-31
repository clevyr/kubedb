package mongodb

import (
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
)

func TestMongoDB_DumpCommand(t *testing.T) {
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
			args{
				&conftypes.Dump{
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				"mongodump",
				"--archive",
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--db=d",
			),
		},
		{
			"clean",
			args{
				&conftypes.Dump{
					Clean:  true,
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				"mongodump",
				"--archive",
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--db=d",
			),
		},
		{
			"tables",
			args{
				&conftypes.Dump{
					Table:  []string{"table1"},
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				"mongodump",
				"--archive",
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--db=d",
				"--collection=table1",
			),
		},
		{
			"exclude-table",
			args{
				&conftypes.Dump{
					ExcludeTable: []string{"table1", "table2"},
					Global:       &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				"mongodump",
				"--archive",
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--db=d",
				"--excludeCollection=table1",
				"--excludeCollection=table2",
			),
		},
		{
			"quiet",
			args{
				&conftypes.Dump{
					Global: &conftypes.Global{
						Host:     "1.1.1.1",
						Database: "d",
						Username: "u",
						Password: "p",
						Quiet:    true,
					},
				},
			},
			command.NewBuilder(
				"mongodump",
				"--archive",
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--db=d",
				"--quiet",
			),
		},
		{
			"port",
			args{&conftypes.Dump{Global: &conftypes.Global{Port: 1234}}},
			command.NewBuilder(
				"mongodump",
				"--archive",
				"--host=",
				"--username=",
				"--authenticationDatabase=",
				"--port=1234",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			got := ma.DumpCommand(tt.args.conf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMongoDB_ExecCommand(t *testing.T) {
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
			args{
				&conftypes.Exec{
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				"exec",
				command.Raw(`"$(which mongosh || which mongo)"`),
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"d",
			),
		},
		{
			"disable-headers",
			args{
				&conftypes.Exec{
					DisableHeaders: true,
					Global:         &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				"exec",
				command.Raw(`"$(which mongosh || which mongo)"`),
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--quiet",
				"d",
			),
		},
		{
			"command",
			args{
				&conftypes.Exec{
					Command: "show databases",
					Global:  &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
			},
			command.NewBuilder(
				"exec",
				command.Raw(`"$(which mongosh || which mongo)"`),
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--eval=show databases",
				"d",
			),
		},
		{
			"port",
			args{&conftypes.Exec{Global: &conftypes.Global{Port: 1234}}},
			command.NewBuilder(
				"exec",
				command.Raw(`"$(which mongosh || which mongo)"`),
				"--host=",
				"--username=",
				"--authenticationDatabase=",
				"--port=1234",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			got := ma.ExecCommand(tt.args.conf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMongoDB_PasswordEnvs(t *testing.T) {
	type args struct {
		c *conftypes.Global
	}
	tests := []struct {
		name string
		args args
		want kubernetes.ConfigLookups
	}{
		{
			"default",
			args{&conftypes.Global{}},
			kubernetes.ConfigLookups{kubernetes.LookupEnv{"MONGODB_EXTRA_PASSWORDS"}},
		},
		{
			"root",
			args{&conftypes.Global{Username: "root"}},
			kubernetes.ConfigLookups{kubernetes.LookupEnv{"MONGODB_ROOT_PASSWORD"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := MongoDB{}
			got := db.PasswordEnvs(tt.args.c)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMongoDB_RestoreCommand(t *testing.T) {
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
			args{
				&conftypes.Restore{
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Gzip,
			},
			command.NewBuilder(
				"mongorestore",
				"--archive",
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--db=d",
			),
		},
		{
			"plain",
			args{
				&conftypes.Restore{
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Plain,
			},
			command.NewBuilder(
				"mongorestore",
				"--archive",
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--db=d",
			),
		},
		{
			"custom",
			args{
				&conftypes.Restore{
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Custom,
			},
			command.NewBuilder(
				"mongorestore",
				"--archive",
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--db=d",
			),
		},
		{
			"clean",
			args{
				&conftypes.Restore{
					Clean:  true,
					Global: &conftypes.Global{Host: "1.1.1.1", Database: "d", Username: "u", Password: "p"},
				},
				sqlformat.Gzip,
			},
			command.NewBuilder(
				"mongorestore",
				"--archive",
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--drop",
				"--db=d",
			),
		},
		{
			"quiet",
			args{
				&conftypes.Restore{
					Global: &conftypes.Global{
						Host:     "1.1.1.1",
						Database: "d",
						Username: "u",
						Password: "p",
						Quiet:    true,
					},
				},
				sqlformat.Gzip,
			},
			command.NewBuilder(
				"mongorestore",
				"--archive",
				"--host=1.1.1.1",
				"--username=u",
				"--authenticationDatabase=d",
				"--password=p",
				"--db=d",
				"--quiet",
			),
		},
		{
			"port",
			args{&conftypes.Restore{Global: &conftypes.Global{Port: 1234}}, sqlformat.Gzip},
			command.NewBuilder(
				"mongorestore",
				"--archive",
				"--host=",
				"--username=",
				"--authenticationDatabase=",
				"--port=1234",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			got := ma.RestoreCommand(tt.args.conf, tt.args.inputFormat)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMongoDB_AuthenticationDatabase(t *testing.T) {
	type args struct {
		c *conftypes.Global
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"root", args{&conftypes.Global{Host: "1.1.1.1", Username: "root"}}, "admin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := MongoDB{}
			got := db.AuthenticationDatabase(tt.args.c)
			assert.Equal(t, tt.want, got)
		})
	}
}
