package config

import (
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/spf13/viper"
)

func TestNamespaceRegexp_Match(t *testing.T) {
	type fields struct {
		re *regexp.Regexp
	}
	type args struct {
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		enabled bool
		want    bool
	}{
		{"dev-ro", fields{regexp.MustCompile(namespaceFilterReadOnly)}, args{"example-dev"}, true, true},
		{"dev-rw", fields{regexp.MustCompile(namespaceFilterReadWrite)}, args{"example-dev"}, true, true},
		{"stage-ro", fields{regexp.MustCompile(namespaceFilterReadOnly)}, args{"example-stage"}, true, true},
		{"stage-rw", fields{regexp.MustCompile(namespaceFilterReadWrite)}, args{"example-stage"}, true, true},
		{"test-ro", fields{regexp.MustCompile(namespaceFilterReadOnly)}, args{"example-test"}, true, true},
		{"test-rw", fields{regexp.MustCompile(namespaceFilterReadWrite)}, args{"example-test"}, true, true},
		{"demo-ro", fields{regexp.MustCompile(namespaceFilterReadOnly)}, args{"example-demo"}, true, true},
		{"demo-rw", fields{regexp.MustCompile(namespaceFilterReadWrite)}, args{"example-demo"}, true, true},
		{"pr-ro", fields{regexp.MustCompile(namespaceFilterReadOnly)}, args{"example-pr1"}, true, true},
		{"pr-rw", fields{regexp.MustCompile(namespaceFilterReadWrite)}, args{"example-pr1"}, true, true},
		{"prod-ro", fields{regexp.MustCompile(namespaceFilterReadOnly)}, args{"example-prod"}, true, true},
		{"prod-rw", fields{regexp.MustCompile(namespaceFilterReadWrite)}, args{"example-prod"}, true, false},
		{"other-ro", fields{regexp.MustCompile(namespaceFilterReadOnly)}, args{"example"}, true, true},
		{"other-rw", fields{regexp.MustCompile(namespaceFilterReadWrite)}, args{"example"}, true, false},
		{"disabled", fields{regexp.MustCompile(namespaceFilterReadWrite)}, args{"example-prod"}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := NamespaceRegexp{
				re: tt.fields.re,
			}
			viper.Set("namespace.filter", tt.enabled)
			if got := level.Match(tt.args.namespace); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewNamespaceRegexp(t *testing.T) {
	type args struct {
		accessLevel string
	}
	tests := []struct {
		name string
		args args
		want NamespaceRegexp
	}{
		{"ro", args{strconv.Itoa(ReadOnly)}, NamespaceRegexp{re: regexp.MustCompile(namespaceFilterReadOnly)}},
		{"rw", args{strconv.Itoa(ReadWrite)}, NamespaceRegexp{re: regexp.MustCompile(namespaceFilterReadWrite)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNamespaceRegexp(tt.args.accessLevel); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNamespaceRegexp() = %v, want %v", got, tt.want)
			}
		})
	}
}
