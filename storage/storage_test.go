package storage_test

import (
	"errors"
	"fmt"
	"testing"
	"github.com/caasmo/kv-repl-barebones/storage"
)

type testCase struct {
	cmd     string
	key     string
	val     string
	want    string
	wantErr error
}

func Example() {

	cases := []struct {
		cmd string
		key string
		val string
	}{
		{cmd: "write", key: "a", val: "hi"},
		{cmd: "read", key: "a", val: ""},
		{cmd: "begin", key: "", val: ""},
		{cmd: "read", key: "a", val: ""},
		{cmd: "write", key: "a", val: "bye"},
		{cmd: "read", key: "a", val: ""},
		{cmd: "begin", key: "", val: ""},
		{cmd: "remove", key: "a", val: ""},
		{cmd: "read", key: "a", val: ""},
		{cmd: "commit", key: "", val: ""},
		{cmd: "read", key: "a", val: ""},
		{cmd: "write", key: "a", val: "bye now"},
		{cmd: "read", key: "a", val: ""},
		{cmd: "discard", key: "", val: ""},
		{cmd: "read", key: "a", val: ""},
	}

	store := storage.NewStore()
	for _, tc := range cases {
		v, err := store.Process(tc.cmd, tc.key, tc.val)
		if len(v) > 0 {
			fmt.Printf("%s\n", v)
		}
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}

	// Output:
	// hi
	// hi
	// bye
	// Key not found: a
	// Key not found: a
	// bye now
	// hi
}

func test(t *testing.T, cases []testCase) {
	store := storage.NewStore()
	for _, tc := range cases {
		v, err := store.Process(tc.cmd, tc.key, tc.val)
		if v != tc.want {
			t.Errorf("\nGot value '%s' want '%s'", v, tc.want)
		}

		if !errors.Is(err, tc.wantErr) {
			t.Errorf("\nGot Error '%s' want '%s'", err, tc.wantErr)
		}
	}
}

func TestWriteRead(t *testing.T) {
	cases := []testCase{
		{cmd: "write", key: "a", val: "hi", want: "", wantErr: nil},
		{cmd: "begin", key: "", val: "", want: "", wantErr: nil},
		{cmd: "write", key: "a", val: "hi 2", want: "", wantErr: nil},
		{cmd: "read", key: "a", val: "", want: "hi 2", wantErr: nil},
		{cmd: "write", key: "a", val: "hi 3", want: "", wantErr: nil},
		{cmd: "read", key: "a", val: "", want: "hi 3", wantErr: nil},
		{cmd: "commit", key: "", val: "", want: "", wantErr: nil},
		{cmd: "read", key: "a", val: "", want: "hi 3", wantErr: nil},
	}

	test(t, cases)
}

func TestRemove(t *testing.T) {
	cases := []testCase{
		{cmd: "remove", key: "a", val: "", want: "", wantErr: storage.ErrKeyNotFound},
		{cmd: "write", key: "a", val: "hi", want: "", wantErr: nil},
		{cmd: "remove", key: "a", val: "", want: "", wantErr: nil},
		{cmd: "remove", key: "a", val: "", want: "", wantErr: storage.ErrKeyNotFound},
	}

	test(t, cases)
}

func TestDiscard(t *testing.T) {
	cases := []testCase{
		{cmd: "write", key: "a", val: "hi", want: "", wantErr: nil},
		{cmd: "begin", key: "", val: "", want: "", wantErr: nil},
		{cmd: "remove", key: "a", val: "", want: "", wantErr: nil},
		{cmd: "discard", key: "", val: "", want: "", wantErr: nil},
		{cmd: "read", key: "a", val: "", want: "hi", wantErr: nil},
	}

	test(t, cases)
}

func TestCommit(t *testing.T) {
	cases := []testCase{
		{cmd: "write", key: "a", val: "hi", want: "", wantErr: nil},
		{cmd: "begin", key: "", val: "", want: "", wantErr: nil},
		{cmd: "write", key: "a", val: "hi 1", want: "", wantErr: nil},
		{cmd: "begin", key: "", val: "", want: "", wantErr: nil},
		{cmd: "write", key: "a", val: "hi 2", want: "", wantErr: nil},
		{cmd: "commit", key: "", val: "", want: "", wantErr: nil},
		{cmd: "commit", key: "", val: "", want: "", wantErr: nil},
		{cmd: "read", key: "a", val: "", want: "hi 2", wantErr: nil},
	}

	test(t, cases)
}

func TestCommitEmpty(t *testing.T) {
	cases := []testCase{
		{cmd: "write", key: "a", val: "hi", want: "", wantErr: nil},
		{cmd: "begin", key: "", val: "", want: "", wantErr: nil},
		{cmd: "begin", key: "", val: "", want: "", wantErr: nil},
		{cmd: "commit", key: "", val: "", want: "", wantErr: nil},
		{cmd: "commit", key: "", val: "", want: "", wantErr: nil},
		{cmd: "read", key: "a", val: "", want: "hi", wantErr: nil},
	}

	test(t, cases)
}

func TestDiscardNoErrorIfNoTransaction(t *testing.T) {
	cases := []testCase{
		{cmd: "write", key: "a", val: "hi", want: "", wantErr: nil},
		{cmd: "discard", key: "", val: "", want: "", wantErr: nil},
	}

	test(t, cases)
}

func TestErrorCommitWithoutTransaction(t *testing.T) {
	cases := []testCase{
		{cmd: "write", key: "a", val: "hi", want: "", wantErr: nil},
		{cmd: "commit", key: "", val: "", want: "", wantErr: storage.ErrNoCurrentTransation},
	}

	test(t, cases)
}

func TestReadErrorNoKey(t *testing.T) {
	cases := []testCase{
		{cmd: "read", key: "a", val: "", want: "", wantErr: storage.ErrKeyNotFound},
	}

	test(t, cases)
}

func TestReadErrorAfterRemoveInKv(t *testing.T) {

	cases := []testCase{
		{cmd: "write", key: "a", val: "hi", want: "", wantErr: nil},
		{cmd: "remove", key: "a", val: "", want: "", wantErr: nil},
		{cmd: "read", key: "a", val: "", want: "", wantErr: storage.ErrKeyNotFound},
	}

	test(t, cases)
}

func TestUnsupportedCommand(t *testing.T) {
	cases := []testCase{
		{cmd: "not-a-supported-command", key: "a", val: "something", want: "", wantErr: storage.ErrUnsupportedCommand},
	}

	test(t, cases)
}

func TestReadErrorAfterRemoveInTransaction(t *testing.T) {

	cases := []testCase{
		{cmd: "write", key: "a", val: "hi", want: "", wantErr: nil},
		{cmd: "begin", key: "", val: "", want: "", wantErr: nil},
		{cmd: "remove", key: "a", val: "", want: "", wantErr: nil},
		{cmd: "read", key: "a", val: "", want: "", wantErr: storage.ErrKeyNotFound},
	}

	test(t, cases)
}
