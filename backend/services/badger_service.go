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
	db         *badger.DB
	defaultTTL time.Duration
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

	ttlStr := utils.AppConfig.CacheTTL
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		ttl = 1 * time.Hour // Default to 1 hour if invalid or not set
	}

	return &BadgerService{db: db, defaultTTL: ttl}, nil
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

// Set stores a key-value pair in the database using the configured default Time-To-Live (TTL).
//
// param key The unique identifier for the data.
// param value The byte array data to store.
// return error An error if the write operation fails.
// @throws error If the transaction fails to commit.
func (s *BadgerService) Set(key string, value []byte) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value).WithTTL(s.defaultTTL)
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
		
		// Debug TTL
		expiresAt := item.ExpiresAt()
		if expiresAt > 0 {
			ttlRemaining := time.Until(time.Unix(int64(expiresAt), 0))
			utils.LogDebug("Cache Hit for '%s' | Expires in: %v", key, ttlRemaining)
		} else {
             // If ExpiresAt is 0, it means the key has no TTL (Persistent)
			utils.LogDebug("Cache Hit for '%s' | Expires in: Never (Persistent)", key)
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

// SetPersistent stores a key-value pair in the database WITHOUT a Time-To-Live (TTL).
// This is used for persistent data that should survive cache flushes, such as device states.
//
// param key The unique identifier for the data.
// param value The byte array data to store.
// return error An error if the write operation fails.
// @throws error If the transaction fails to commit.
func (s *BadgerService) SetPersistent(key string, value []byte) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		// No TTL - data persists indefinitely
		return txn.Set([]byte(key), value)
	})
	if err != nil {
		utils.LogError("BadgerService: failed to set persistent key %s: %v", key, err)
		return err
	}
	utils.LogDebug("BadgerService: Set persistent key '%s' (no TTL)", key)
	return nil
}

// GetAllKeysWithPrefix retrieves all keys that start with the specified prefix.
// This is useful for cleanup operations or listing related items.
//
// param prefix The string pattern to match at the beginning of keys.
// return []string A slice of all matching keys.
// return error An error if the iteration fails.
func (s *BadgerService) GetAllKeysWithPrefix(prefix string) ([]string, error) {
	var keys []string
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need keys, not values
		it := txn.NewIterator(opts)
		defer it.Close()

		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := string(item.Key())
			keys = append(keys, key)
		}
		return nil
	})

	if err != nil {
		utils.LogError("BadgerService: failed to get keys with prefix %s: %v", prefix, err)
		return nil, err
	}

	utils.LogDebug("BadgerService: Found %d keys with prefix '%s'", len(keys), prefix)
	return keys, nil
}

// FlushAll removes all CACHE data from the database (keys with "cache:" prefix).
// Device state and other persistent data (without "cache:" prefix) are preserved.
// This is a selective flush operation, not a complete database wipe.
//
// return error An error if the drop operation fails.
func (s *BadgerService) FlushAll() error {
	// Only clear keys with "cache:" prefix
	cachePrefix := "cache:"
	err := s.db.DropPrefix([]byte(cachePrefix))
	if err != nil {
		utils.LogError("BadgerService: failed to flush cache: %v", err)
		return err
	}
	utils.LogInfo("BadgerService: Flushed all cache data (preserved persistent data)")
	return nil
}
