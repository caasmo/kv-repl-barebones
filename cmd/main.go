package main

import (
	"github.com/caasmo/kv-repl-barebones/repl"
	"github.com/caasmo/kv-repl-barebones/storage"
)

func main() {
	store := storage.NewStore()
	repl.NewRepl(store).Run()
}
