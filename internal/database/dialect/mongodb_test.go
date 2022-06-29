package dialect

import (
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func TestMongoDB_AnalyzeQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.AnalyzeQuery(); got != tt.want {
				t.Errorf("AnalyzeQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDB_DatabaseEnvNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"default", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.DatabaseEnvNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseEnvNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDB_DefaultPort(t *testing.T) {
	tests := []struct {
		name string
		want uint16
	}{
		{"default", 27017},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.DefaultPort(); got != tt.want {
				t.Errorf("DefaultPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDB_DefaultUser(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "root"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.DefaultUser(); got != tt.want {
				t.Errorf("DefaultUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDB_DropDatabaseQuery(t *testing.T) {
	type args struct {
		database string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"database", args{"database"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.DropDatabaseQuery(tt.args.database); got != tt.want {
				t.Errorf("DropDatabaseQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			args{config.Dump{Global: config.Global{Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("mongodump", "--archive", "--host=127.0.0.1", "--username=u", "--password=p", "--db=d"),
		},
		{
			"clean",
			args{config.Dump{Clean: true, Global: config.Global{Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("mongodump", "--archive", "--host=127.0.0.1", "--username=u", "--password=p", "--db=d"),
		},
		{
			"tables",
			args{config.Dump{Tables: []string{"table1"}, Global: config.Global{Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("mongodump", "--archive", "--host=127.0.0.1", "--username=u", "--password=p", "--db=d", "--collection=table1"),
		},
		{
			"exclude-table",
			args{config.Dump{ExcludeTable: []string{"table1", "table2"}, Global: config.Global{Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("mongodump", "--archive", "--host=127.0.0.1", "--username=u", "--password=p", "--db=d", "--excludeCollection=table1", "--excludeCollection=table2"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.DumpCommand(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DumpCommand() = %v, want %v", got, tt.want)
			}
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
			args{config.Exec{Global: config.Global{Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("mongosh", "--host=127.0.0.1", "--username=u", "--password=p", "d"),
		},
		{
			"disable-headers",
			args{config.Exec{DisableHeaders: true, Global: config.Global{Database: "d", Username: "u", Password: "p"}}},
			command.NewBuilder("mongosh", "--host=127.0.0.1", "--username=u", "--password=p", "--quiet", "d"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.ExecCommand(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExecCommand() = %v, want %v", got, tt.want)
			}
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
			got, err := ma.FilterPods(tt.args.client, tt.args.pods)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterPods() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDB_ListDatabasesQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "db.getMongo().getDBNames().forEach(function(db){ print(db) })"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.ListDatabasesQuery(); got != tt.want {
				t.Errorf("ListDatabasesQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDB_ListTablesQuery(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "show collections"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.ListTablesQuery(); got != tt.want {
				t.Errorf("ListTablesQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDB_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", "mongodb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDB_PasswordEnvNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"default", []string{"MONGODB_ROOT_PASSWORD"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.PasswordEnvNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PasswordEnvNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDB_PodLabels(t *testing.T) {
	tests := []struct {
		name string
		want []kubernetes.LabelQueryable
	}{
		{"default", []kubernetes.LabelQueryable{
			kubernetes.LabelQuery{Name: "app", Value: "mongodb"},
			kubernetes.LabelQueryAnd{
				{Name: "app.kubernetes.io/name", Value: "mongodb"},
				{Name: "app.kubernetes.io/component", Value: "mongodb"},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.PodLabels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PodLabels() = %v, want %v", got, tt.want)
			}
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
			args{config.Restore{Global: config.Global{Database: "d", Username: "u", Password: "p"}}, sqlformat.Gzip},
			command.NewBuilder("mongorestore", "--archive", "--host=127.0.0.1", "--username=u", "--password=p", "--db=d"),
		},
		{
			"plain",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u", Password: "p"}}, sqlformat.Plain},
			command.NewBuilder("mongorestore", "--archive", "--host=127.0.0.1", "--username=u", "--password=p", "--db=d"),
		},
		{
			"custom",
			args{config.Restore{Global: config.Global{Database: "d", Username: "u", Password: "p"}}, sqlformat.Custom},
			command.NewBuilder("mongorestore", "--archive", "--host=127.0.0.1", "--username=u", "--password=p", "--db=d"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.RestoreCommand(tt.args.conf, tt.args.inputFormat); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RestoreCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDB_UserEnvNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"default", []string{"MONGODB_ROOT_USER"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := MongoDB{}
			if got := ma.UserEnvNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserEnvNames() = %v, want %v", got, tt.want)
			}
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
			if got := po.DumpExtension(tt.args.format); got != tt.want {
				t.Errorf("DumpFileExt() = %v, want %v", got, tt.want)
			}
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
			if got := ma.FormatFromFilename(tt.args.filename); got != tt.want {
				t.Errorf("FormatFromFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
