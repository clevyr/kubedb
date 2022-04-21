package command

import (
	"reflect"
	"testing"
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
			if got := e.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
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
			if got := NewEnv(tt.args.k, tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
