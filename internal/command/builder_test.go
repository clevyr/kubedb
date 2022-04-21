package command

import (
	"reflect"
	"testing"
)

func TestBuilder_Push(t *testing.T) {
	type fields struct {
		cmd []any
	}
	type args struct {
		p []any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Builder
	}{
		{"empty", fields{[]any{}}, args{[]any{"echo", "hello world"}}, &Builder{[]any{"echo", "hello world"}}},
		{"append", fields{[]any{"echo"}}, args{[]any{"hello world"}}, &Builder{[]any{"echo", "hello world"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &Builder{
				cmd: tt.fields.cmd,
			}
			if got := j.Push(tt.args.p...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Push() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_String(t *testing.T) {
	type fields struct {
		cmd []any
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{"simple", fields{[]any{"echo", "hello", "world"}}, "echo hello world", false},
		{"pipe", fields{[]any{"echo", "hello", "world", Pipe, "cat"}}, "echo hello world | cat", false},
		{"escape", fields{[]any{"echo", "hello world"}}, "echo 'hello world'", false},
		{"env", fields{[]any{Env{"MESSAGE", "hello world"}, "env"}}, Env{"MESSAGE", "hello world"}.String() + " env", false},
		{"panic", fields{[]any{0}}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := recover(); (err != nil) != tt.wantPanic {
					t.Errorf("String() panic = %v, wantPanic %v", err, tt.wantPanic)
				}
			}()
			j := Builder{
				cmd: tt.fields.cmd,
			}
			if got := j.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_Unshift(t *testing.T) {
	type fields struct {
		cmd []any
	}
	type args struct {
		p []any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Builder
	}{
		{"empty", fields{[]any{}}, args{[]any{"echo", "hello", "world"}}, &Builder{[]any{"echo", "hello", "world"}}},
		{"prepend", fields{[]any{"hello", "world"}}, args{[]any{"echo"}}, &Builder{[]any{"echo", "hello", "world"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &Builder{
				cmd: tt.fields.cmd,
			}
			if got := j.Unshift(tt.args.p...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unshift() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBuilder(t *testing.T) {
	type args struct {
		p []any
	}
	tests := []struct {
		name string
		args args
		want *Builder
	}{
		{"empty", args{[]any{}}, &Builder{[]any{}}},
		{"simple", args{[]any{"echo", "hello world"}}, &Builder{[]any{"echo", "hello world"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBuilder(tt.args.p...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}
