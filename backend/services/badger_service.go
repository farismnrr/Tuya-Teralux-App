package services

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
	"teralux_app/utils"
)

// BadgerService handles BadgerDB operations for caching and data persistence.
// It wraps the raw BadgerDB client to provide simplified methods for common operations.
type BadgerService struct {
	db *badger.DB
}

// NewBadgerService initializes a new BadgerService instance.
//
// param dbPath rule="required" The file system path where the database directory will be created or opened.
// return *BadgerService A pointer to the initialized service instance ready for use.
// return error An error if the database cannot be opened (e.g., permissions, locked).
// @throws error If BadgerDB fails to open the database file.
func NewBadgerService(dbPath string) (*BadgerService, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &BadgerService{db: db}, nil
}

// Close terminates the database connection and ensures all data is flushed to disk.
// This method should be called ensuring graceful shutdown of the application.
//
// return error An error if the closing process encounters any issue.
func (s *BadgerService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Set stores a key-value pair in the database with a specified Time-To-Live (TTL).
//
// param key The unique identifier for the data.
// param value The byte array data to store.
// param ttl The duration after which the key should expire.
// return error An error if the write operation fails.
// @throws error If the transaction fails to commit.
func (s *BadgerService) Set(key string, value []byte, ttl time.Duration) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value).WithTTL(ttl)
		return txn.SetEntry(entry)
	})
	if err != nil {
		utils.LogError("BadgerService: failed to set key %s: %v", key, err)
		return err
	}
	return nil
}

// Get retrieves a value associated with the given key.
// It handles the transaction view automatically.
//
// param key The unique identifier to search for.
// return []byte The value stored under the key, or nil if the key does not exist.
// return error An error if the read operation fails (excluding KeyNotFound).
// @throws error if an internal database error occurs during the view transaction.
func (s *BadgerService) Get(key string) ([]byte, error) {
	var valCopy []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		valCopy, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil // Return nil if not found, distinct from error
		}
		utils.LogError("BadgerService: failed to get key %s: %v", key, err)
		return nil, err
	}

	return valCopy, nil
}

// Delete removes a key and its associated value from the database.
//
// param key The unique identifier to remove.
// return error An error if the delete operation fails.
// @throws error If the transaction fails to commit.
func (s *BadgerService) Delete(key string) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil {
		utils.LogError("BadgerService: failed to delete key %s: %v", key, err)
		return err
	}
	return nil
}

// ClearWithPrefix removes all keys that start with the specified prefix.
// This is useful for clearing a group of related cache items.
//
// param prefix The string pattern to match at the beginning of keys.
// return error An error if the bulk drop operation fails.
func (s *BadgerService) ClearWithPrefix(prefix string) error {
	return s.db.DropPrefix([]byte(prefix))
}

// FlushAll removes all data from the database.
// WARNING: This action is irreversible.
//
// return error An error if the drop all operation fails.
func (s *BadgerService) FlushAll() error {
	return s.db.DropAll()
}
