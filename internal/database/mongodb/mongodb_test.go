package mongodb

import (
	"context"
	"testing"

	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
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

func TestMongoDB_FilterPods(t *testing.T) {
	type args struct {
		client kubernetes.KubeClient
		pods   []v1.Pod
	}
	tests := []struct {
		name    string
		args    args
		want    []v1.Pod
		wantErr bool
	}{
		{"empty", args{kubernetes.KubeClient{}, []v1.Pod{}}, []v1.Pod{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			got, err := ma.FilterPods(context.TODO(), tt.args.client, tt.args.pods)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

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
		want kubernetes.ConfigFinders
	}{
		{"default", args{}, kubernetes.ConfigFinders{kubernetes.ConfigFromEnv{"MONGODB_EXTRA_PASSWORDS"}}},
		{"root", args{config.Global{Username: "root"}}, kubernetes.ConfigFinders{kubernetes.ConfigFromEnv{"MONGODB_ROOT_PASSWORD"}}},
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

func TestMongoDB_DumpExtension(t *testing.T) {
	type args struct {
		format sqlformat.Format
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"plain", args{sqlformat.Plain}, ".archive"},
		{"gzip", args{sqlformat.Gzip}, ".archive.gz"},
		{"custom", args{sqlformat.Custom}, ""},
		{"unknown", args{sqlformat.Unknown}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := MongoDB{}
			got := po.DumpExtension(tt.args.format)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMongoDB_FormatFromFilename(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want sqlformat.Format
	}{
		{"gzip", args{"test.archive.gz"}, sqlformat.Gzip},
		{"plain", args{"test.archive"}, sqlformat.Plain},
		{"unknown", args{"test.png"}, sqlformat.Unknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			got := ma.FormatFromFilename(tt.args.filename)
			assert.Equal(t, tt.want, got)
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
