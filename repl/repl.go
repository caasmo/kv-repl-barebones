// Package repl implements a simple repl (Read-Eval-Print Loop) for access to a
// Key Value Storage system.
package repl

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/caasmo/kv-repl-barebones/storage"
	"os"
	"strings"
)

// exit is the command to exit the repl
const exit = "exit"

// validCommands are the commands supported by the repl
//
// The values of the map are the required number of arguments for each command.
var validCommands = map[string]int{
	storage.Write:   2,
	storage.Read:    1,
	storage.Remove:  1,
	storage.Begin:   0,
	storage.Commit:  0,
	storage.Discard: 0,
	exit:            0,
}

var (
	errUnsupportedCommand  error = errors.New("Unsupported command")
	errNoCommand           error = errors.New("No command given")
	errInvalidNumArguments error = errors.New("Invalid Number of arguments")
)

// repl represents a simple repl (Read, Evaluate, Print and Loop).
type repl struct {
	store *storage.Store
}

// NewRepl returns a repl.
func NewRepl(s *storage.Store) *repl {
	return &repl{store: s}
}

// prompt prints the prompt to Stdout.
func (r *repl) prompt() {
	fmt.Print("> ")
}

// read reads the user input from the os.Stdin
func (r *repl) read() string {
	reader := bufio.NewReader(os.Stdin)
	t, _ := reader.ReadString('\n')
	return strings.TrimSpace(t)
}

// print prints a string to Stdout.
func (r *repl) print(msg string) {
	fmt.Println(msg)
}

// parse parses and validates the input from the user.
// It returns the command, key, value and error.
func (r *repl) parse(in string) (string, string, string, error) {

	// Commands are case-insensitive.
	in = strings.ToLower(in)

	fields := strings.Fields(in)

	if len(fields) == 0 {
		return "", "", "", errNoCommand
	}

	numParams, ok := validCommands[fields[0]]

	if !ok {
		return "", "", "", fmt.Errorf("%w: %s", errUnsupportedCommand, fields[0])
	}

	if numParams != len(fields)-1 {
		return "", "", "", fmt.Errorf("%w: %s (required: %d)", errInvalidNumArguments, strings.ToUpper(fields[0]), numParams)
	}

	command := fields[0]
	key := ""
	value := ""

	switch len(fields) {
	case 1:
		break
	case 2:
		key = fields[1]
		break
	case 3:
		key = fields[1]
		value = fields[2]
	}

	return command, key, value, nil
}

// Run starts the repl.
func (r repl) Run() {
	for {
		r.next()
	}
}

// next iterates the repl.
func (r repl) next() {
	r.prompt()
	in := r.read()
	cmd, key, value, err := r.parse(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// exit is a repl command, not a storage one. Handled here.
	if cmd == exit {
		os.Exit(0)
	}

	v, err := r.store.Process(cmd, key, value)

	// All errors are output to stderr.
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// For simpicity empty values are not allowed.
	if len(v) > 0 {
		fmt.Fprintln(os.Stdout, v)
	}

	return
}
