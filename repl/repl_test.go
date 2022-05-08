package repl

import (
	"errors"
	"github.com/caasmo/kv-repl-barebones/storage"
	"testing"
)

func TestParseErrors(t *testing.T) {
	store := &storage.Store{}
	r := NewRepl(store)

	cases := []struct {
		input   string
		wantErr error
	}{
		{input: "", wantErr: errNoCommand},
		{input: "unsupported-comand", wantErr: errUnsupportedCommand},
		{input: "write 1 2", wantErr: nil},
		{input: "write", wantErr: errInvalidNumArguments},
		{input: "write 1 2 3 ", wantErr: errInvalidNumArguments},
		{input: "read a", wantErr: nil},
		{input: "read a b", wantErr: errInvalidNumArguments},
		{input: "read", wantErr: errInvalidNumArguments},
		{input: "remove a", wantErr: nil},
		{input: "remove", wantErr: errInvalidNumArguments},
		{input: "remove a b", wantErr: errInvalidNumArguments},
		{input: "discard", wantErr: nil},
		{input: "discard 4", wantErr: errInvalidNumArguments},
		{input: "begin", wantErr: nil},
		{input: "begin 4", wantErr: errInvalidNumArguments},
		{input: "commit", wantErr: nil},
		{input: "commit 4", wantErr: errInvalidNumArguments},
		{input: "exit", wantErr: nil},
		{input: "exit 4", wantErr: errInvalidNumArguments},
	}

	for _, tc := range cases {
		_, _, _, err := r.parse(tc.input)

		if !errors.Is(err, tc.wantErr) {
			t.Errorf("\nGot Error '%s' want '%s'", err, tc.wantErr)
		}
	}
}

func TestParse(t *testing.T) {
	wantCmd := "write"
	wantKey := "a"
	wantVal := "hi"

	store := &storage.Store{}
	r := NewRepl(store)

	cmd, key, val, err := r.parse("write a hi")

	if err != nil {
		t.Errorf("\nGot Error '%s' want 'nil'", err)
	}

	if cmd != wantCmd {
		t.Errorf("\nGot cmd '%s' want '%s'", cmd, wantCmd)
	}

	if key != wantKey {
		t.Errorf("\nGot key '%s' want '%s'", key, wantKey)
	}

	if val != wantVal {
		t.Errorf("\nGot cmd '%s' want '%s'", val, wantVal)
	}
}
