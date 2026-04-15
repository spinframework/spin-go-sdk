// Package kv provides access to Spin key-value stores.
package kv

import (
	"fmt"

	keyvalue "github.com/spinframework/spin-go-sdk/v3/imports/fermyon_spin_2_0_0_key_value"
)

// Store represents a connection to a key-value store.
type Store struct {
	store *keyvalue.Store
}

// Open opens the store with the specified label.
func Open(label string) (*Store, error) {
	result := keyvalue.StoreOpen(label)
	if result.IsErr() {
		return nil, errorVariantToError(result.Err())
	}

	return &Store{
		store: result.Ok(),
	}, nil
}

// OpenDefault opens the default store.
//
// This is equivalent to Open("default").
func OpenDefault() (*Store, error) {
	return Open("default")
}

// Set sets the key/value pair in the store.
func (s *Store) Set(key string, value []byte) error {
	result := s.store.Set(key, value)
	if result.IsErr() {
		return errorVariantToError(result.Err())
	}

	return nil
}

// Get returns the value of the provided key from the store.
func (s *Store) Get(key string) ([]byte, error) {
	result := s.store.Get(key)
	if result.IsErr() {
		return nil, errorVariantToError(result.Err())
	}

	value := result.Ok()
	if value.IsNone() {
		return []byte(""), nil
	}

	return value.Some(), nil
}

// Delete removes the given key/value from the store.
func (s *Store) Delete(key string) error {
	result := s.store.Delete(key)
	if result.IsErr() {
		return errorVariantToError(result.Err())
	}

	return nil
}

// Exists checks if a given key exists in the store.
func (s *Store) Exists(key string) (bool, error) {
	result := s.store.Exists(key)
	if result.IsErr() {
		return false, errorVariantToError(result.Err())
	}

	return result.Ok(), nil
}

// GetKeys returns all the keys from the store.
func (s *Store) GetKeys() ([]string, error) {
	result := s.store.GetKeys()
	if result.IsErr() {
		return nil, errorVariantToError(result.Err())
	}

	return result.Ok(), nil
}

func errorVariantToError(code keyvalue.Error) error {
	switch code.Tag() {
	case keyvalue.ErrorAccessDenied:
		return fmt.Errorf("access denied")
	case keyvalue.ErrorNoSuchStore:
		return fmt.Errorf("no such store")
	case keyvalue.ErrorStoreTableFull:
		return fmt.Errorf("store table full")
	case keyvalue.ErrorOther:
		return fmt.Errorf("%v", code.Other())
	default:
		return fmt.Errorf("no error provided by host implementation")
	}
}
