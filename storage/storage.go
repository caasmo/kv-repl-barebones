// Package storage implements a simple Key Value Storage system with nested
// transaction capabilities.
package storage

import (
	"errors"
	"fmt"
)

const (
	// Supported commands
	Write   = "write"
	Read    = "read"
	Remove  = "remove"
	Begin   = "begin"
	Commit  = "commit"
	Discard = "discard"
)

var (
	ErrNoCurrentTransation error = errors.New("There is no current transaction to commit")
	ErrKeyNotFound         error = errors.New("Key not found")
	ErrUnsupportedCommand  error = errors.New("Unsupported command")
)

// operation represents a unit of a transaction. An operation modifies
// eventually the state of the kv. operations are appended to the transaction
// or written in the kv sequencially. An operation can only modify the state of
// the kv by writing (isWrite = true) or removing (isWrite = false).
type operation struct {
	key     string
	value   string
	isWrite bool
}

// tx represents a transaction. A transaction has a parent transaction. All
// operations of a transaction are "eventually" commited to the parent
// transaction or to the the kv store if there is no parent.
type tx struct {
	parent     *tx
	operations []operation
}

// isRoot returns true if the transaction tx has no parent.
func (t *tx) isRoot() bool {
	if t.parent == nil {
		return true
	}

	return false
}

// hasOperations returns true if the transaction tx contains operations.
func (t *tx) hasOperations() bool {
	if len(t.operations) > 0 {
		return true
	}

	return false
}

// kvStore represents a in-memory Key Value storage system.
//
// As this is just a barebones kv Store for one client, there is no need for
// locking or multiple threads.
type kvStore map[string]string

// modify applies an operation to the kvStore. Depending on the isWrite flag
// of the operation op, modify writes or removes to the kvStore.
func (kv kvStore) modify(op operation) {
	switch op.isWrite {
	case true:
		kv[op.key] = op.value
	case false:
		delete(kv, op.key)
	}
}

// A Store represents a key value storage system with transaction capabilities.
// A Store contains the kvStore and a pointer to the data that can eventually
// be commited to the kvStore (currTx).
type Store struct {
	kv     kvStore
	currTx *tx
}

// Process processes a command.
//
// It returns an error if the Store does not support the command.
// It also returns an error if the Store rejects the command.
func (s *Store) Process(command, key, value string) (string, error) {

	switch command {
	case Write:
		s.write(key, value)
		return "", nil
	case Read:
		return s.read(key)
	case Remove:
		return "", s.remove(key)
	case Begin:
		s.begin()
		return "", nil
	case Discard:
		s.discard()
		return "", nil
	case Commit:
		return "", s.commit()
	}

	return "", fmt.Errorf("%w: %s", ErrUnsupportedCommand, command)
}

// modify applies the operation op to the Store. modify either writes to the
// kvStore or appends the operation to the current transaction.
func (s *Store) modify(op operation) {
	if s.currTx.isRoot() {
		//write db
		s.kv.modify(op)
	} else {
		// append to transaction operations
		s.currTx.operations = append(s.currTx.operations, op)
	}
}

// NewStore returns a Store.
func NewStore() *Store {
	return &Store{kv: make(map[string]string), currTx: &tx{}}
}

// write writes the value and the key to the Store. Depending of the current
// transaction, it writes to the kvStore or the to current transation.
func (s *Store) write(key, value string) {
	s.modify(operation{key: key, value: value, isWrite: true})
}

// read retrieves the current value of the key key. The value can be on the
// transaction or already written in the kvStore.
//
// read returns error if the key does not exist.
func (s *Store) read(key string) (string, error) {
	currentTx := s.currTx
	for !currentTx.isRoot() {
		// search for the key recursively and in reverse
		for i := len(s.currTx.operations) - 1; i >= 0; i-- {
			if key == s.currTx.operations[i].key {

				// false means key was deleted in the transaction
				if false == s.currTx.operations[i].isWrite {
					return "", fmt.Errorf("%w: %s", ErrKeyNotFound, key)
				}

				return s.currTx.operations[i].value, nil
			}
		}

		currentTx = currentTx.parent
	}

	// the key is not in the transactions. Check the kv
	v, ok := s.kv[key]
	if ok {
		return v, nil
	}

	return "", fmt.Errorf("%w: %s", ErrKeyNotFound, key)
}

// remove removes the key from the kvStore, or marks the key for removal in the
// current transaction.
//
// remove returns error if the key does not exist.
func (s *Store) remove(key string) error {

	_, err := s.read(key)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	s.modify(operation{key: key, isWrite: false})
	return nil
}

// commit applies all operations of the curent transaction to the parent
// transaction or to the kvStore if the transaction has no parent.
func (s *Store) commit() error {

	if s.currTx.isRoot() {
		return ErrNoCurrentTransation
	}

	// 1) append to parent
	for _, op := range s.currTx.operations {
		s.currTx.parent.operations = append(s.currTx.parent.operations, operation{key: op.key, value: op.value, isWrite: op.isWrite})
	}

	// 2) delete/sustitute current
	s.currTx = s.currTx.parent

	// 3) if new current parent is root and has operations is, apply them
    // sequentially. No intend is made to optimize the operations. F. ex, only
    // apply the last write for each key.
	if s.currTx.isRoot() && s.currTx.hasOperations() {
		for _, op := range s.currTx.operations {
			s.kv.modify(op)
		}

		// delete the operations, as they are now in the kvStore
		s.currTx.operations = nil
	}

	return nil
}

// discard discards the current transaction. All operations in the current
// transaction are discarded.
func (s *Store) discard() {

	if s.currTx.isRoot() {
		return
	}

	s.currTx = s.currTx.parent
}

// begin initiates a transaction.
func (s *Store) begin() {
	s.currTx = &tx{parent: s.currTx}
}
