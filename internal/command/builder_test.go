package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type panicFunc func(assert.TestingT, assert.PanicTestFunc, ...any) bool

func TestBuilder_Push(t *testing.T) {
	type fields struct {
		cmd []any
	}
	type args struct {
		p []any
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *Builder
		wantPanic panicFunc
	}{
		{
			"empty",
			fields{[]any{}},
			args{[]any{"echo", "hello world"}},
			&Builder{[]any{"echo", "hello world"}},
			assert.NotPanics,
		},
		{
			"append",
			fields{[]any{"echo"}},
			args{[]any{"hello world"}},
			&Builder{[]any{"echo", "hello world"}},
			assert.NotPanics,
		},
		{"panic", fields{}, args{[]any{0}}, nil, assert.Panics},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantPanic(t, func() {
				j := &Builder{
					cmd: tt.fields.cmd,
				}
				got := j.Push(tt.args.p...)
				assert.Equal(t, tt.want, got)
			})
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
		wantPanic panicFunc
	}{
		{"simple", fields{[]any{"echo", "hello", "world"}}, "echo hello world", assert.NotPanics},
		{"pipe", fields{[]any{"echo", "hello", "world", Pipe, "cat"}}, "echo hello world | cat", assert.NotPanics},
		{"escape", fields{[]any{"echo", "hello world"}}, "echo 'hello world'", assert.NotPanics},
		{
			"env",
			fields{[]any{Env{"MESSAGE", "hello world"}, "env"}},
			Env{"MESSAGE", "hello world"}.Quote() + " env",
			assert.NotPanics,
		},
		{"panic", fields{[]any{0}}, "", assert.Panics},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantPanic(t, func() {
				j := Builder{
					cmd: tt.fields.cmd,
				}
				got := j.String()
				assert.Equal(t, tt.want, got)
			})
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
		name      string
		fields    fields
		args      args
		want      *Builder
		wantPanic panicFunc
	}{
		{
			"empty",
			fields{[]any{}},
			args{[]any{"echo", "hello", "world"}},
			&Builder{[]any{"echo", "hello", "world"}},
			assert.NotPanics,
		},
		{
			"prepend",
			fields{[]any{"hello", "world"}},
			args{[]any{"echo"}},
			&Builder{[]any{"echo", "hello", "world"}},
			assert.NotPanics,
		},
		{"panic", fields{}, args{[]any{0}}, nil, assert.Panics},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantPanic(t, func() {
				j := &Builder{
					cmd: tt.fields.cmd,
				}
				got := j.Unshift(tt.args.p...)
				assert.Equal(t, tt.want, got)
			})
		})
	}
}

func TestNewBuilder(t *testing.T) {
	type args struct {
		p []any
	}
	tests := []struct {
		name      string
		args      args
		want      *Builder
		wantPanic panicFunc
	}{
		{"empty", args{[]any{}}, &Builder{[]any{}}, assert.NotPanics},
		{"simple", args{[]any{"echo", "hello world"}}, &Builder{[]any{"echo", "hello world"}}, assert.NotPanics},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBuilder(tt.args.p...)
			assert.Equal(t, tt.want, got)
		})
	}
}
