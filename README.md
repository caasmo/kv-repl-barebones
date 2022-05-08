## kv-repl-barebones

**kv-repl-barebones** is a simple exercise to show a barebones Read-Eval-Print
Loop, or REPL, for managing a basic key-value store.

## Run

    go run cmd/main.go
    > read k
    Key not found: k
    > begin
    > write k 42
    > read k
    42
    > commit
    > read k
    42
    > begin
    > remove k
    > read k
    Key not found: k
    > discard
    > read k
    42
    > exit

## Test

    go test -v -coverprofile=c.out ./...

