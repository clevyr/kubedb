package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			got := j.Push(tt.args.p...)
			assert.Equal(t, tt.want, got)
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
		{"env", fields{[]any{Env{"MESSAGE", "hello world"}, "env"}}, Env{"MESSAGE", "hello world"}.Quote() + " env", false},
		{"panic", fields{[]any{0}}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				err, ok := recover().(error)
				if tt.wantPanic {
					assert.True(t, ok)
					assert.Error(t, err)
				} else {
					assert.False(t, ok)
					assert.NoError(t, err)
				}
			}()
			j := Builder{
				cmd: tt.fields.cmd,
			}
			got := j.String()
			assert.Equal(t, tt.want, got)
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
			got := j.Unshift(tt.args.p...)
			assert.Equal(t, tt.want, got)
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
			got := NewBuilder(tt.args.p...)
			assert.Equal(t, tt.want, got)
		})
	}
}
