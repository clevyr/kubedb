package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv_String(t *testing.T) {
	type fields struct {
		Key   string
		Value string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"simple", fields{"MESSAGE", "hello"}, "MESSAGE=hello"},
		{"escaped", fields{"MESSAGE", "hello world"}, "MESSAGE='hello world'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Env{
				Key:   tt.fields.Key,
				Value: tt.fields.Value,
			}
			got := e.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewEnv(t *testing.T) {
	type args struct {
		k string
		v string
	}
	tests := []struct {
		name string
		args args
		want Env
	}{
		{"simple", args{"MESSAGE", "hello world"}, Env{"MESSAGE", "hello world"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewEnv(tt.args.k, tt.args.v)
			assert.Equal(t, tt.want, got)
		})
	}
}
