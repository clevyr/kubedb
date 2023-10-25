package namespace_filter

import (
	"context"
	"regexp"
	"testing"

	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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
			viper.Set(consts.NamespaceFilterKey, tt.enabled)
			got := level.Match(tt.args.namespace)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewFromContext(t *testing.T) {
	roCtx := NewContext(context.Background(), ReadOnly)
	rwCtx := NewContext(context.Background(), ReadWrite)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want NamespaceRegexp
	}{
		{"ro", args{roCtx}, NamespaceRegexp{re: regexp.MustCompile(namespaceFilterReadOnly)}},
		{"rw", args{rwCtx}, NamespaceRegexp{re: regexp.MustCompile(namespaceFilterReadWrite)}},
		{"unset", args{context.Background()}, NamespaceRegexp{re: regexp.MustCompile(namespaceFilterReadWrite)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewFromContext(tt.args.ctx))
		})
	}
}