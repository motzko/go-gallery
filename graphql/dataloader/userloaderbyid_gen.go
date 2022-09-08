// Code generated by github.com/vektah/dataloaden, DO NOT EDIT.

package dataloader

import (
	"sync"
	"time"

	"github.com/mikeydub/go-gallery/db/gen/coredb"
	"github.com/mikeydub/go-gallery/service/persist"
)

// UserLoaderByIDConfig captures the config to create a new UserLoaderByID
type UserLoaderByIDConfig struct {
	// Fetch is a method that provides the data for the loader
	Fetch func(keys []persist.DBID) ([]coredb.User, []error)

	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

// NewUserLoaderByID creates a new UserLoaderByID given a fetch, wait, and maxBatch
func NewUserLoaderByID(config UserLoaderByIDConfig) *UserLoaderByID {
	return &UserLoaderByID{
		fetch:    config.Fetch,
		wait:     config.Wait,
		maxBatch: config.MaxBatch,
	}
}

// UserLoaderByID batches and caches requests
type UserLoaderByID struct {
	// this method provides the data for the loader
	fetch func(keys []persist.DBID) ([]coredb.User, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// INTERNAL

	// lazily created cache
	cache map[persist.DBID]coredb.User

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *userLoaderByIDBatch

	// mutex to prevent races
	mu sync.Mutex
}

type userLoaderByIDBatch struct {
	keys    []persist.DBID
	data    []coredb.User
	error   []error
	closing bool
	done    chan struct{}
}

// Load a User by key, batching and caching will be applied automatically
func (l *UserLoaderByID) Load(key persist.DBID) (coredb.User, error) {
	return l.LoadThunk(key)()
}

// LoadThunk returns a function that when called will block waiting for a User.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *UserLoaderByID) LoadThunk(key persist.DBID) func() (coredb.User, error) {
	l.mu.Lock()
	if it, ok := l.cache[key]; ok {
		l.mu.Unlock()
		return func() (coredb.User, error) {
			return it, nil
		}
	}
	if l.batch == nil {
		l.batch = &userLoaderByIDBatch{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(l, key)
	l.mu.Unlock()

	return func() (coredb.User, error) {
		<-batch.done

		var data coredb.User
		if pos < len(batch.data) {
			data = batch.data[pos]
		}

		var err error
		// its convenient to be able to return a single error for everything
		if len(batch.error) == 1 {
			err = batch.error[0]
		} else if batch.error != nil {
			err = batch.error[pos]
		}

		if err == nil {
			l.mu.Lock()
			l.unsafeSet(key, data)
			l.mu.Unlock()
		}

		return data, err
	}
}

// LoadAll fetches many keys at once. It will be broken into appropriate sized
// sub batches depending on how the loader is configured
func (l *UserLoaderByID) LoadAll(keys []persist.DBID) ([]coredb.User, []error) {
	results := make([]func() (coredb.User, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}

	users := make([]coredb.User, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		users[i], errors[i] = thunk()
	}
	return users, errors
}

// LoadAllThunk returns a function that when called will block waiting for a Users.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *UserLoaderByID) LoadAllThunk(keys []persist.DBID) func() ([]coredb.User, []error) {
	results := make([]func() (coredb.User, error), len(keys))
	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}
	return func() ([]coredb.User, []error) {
		users := make([]coredb.User, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			users[i], errors[i] = thunk()
		}
		return users, errors
	}
}

// Prime the cache with the provided key and value. If the key already exists, no change is made
// and false is returned.
// (To forcefully prime the cache, clear the key first with loader.clear(key).prime(key, value).)
func (l *UserLoaderByID) Prime(key persist.DBID, value coredb.User) bool {
	l.mu.Lock()
	var found bool
	if _, found = l.cache[key]; !found {
		l.unsafeSet(key, value)
	}
	l.mu.Unlock()
	return !found
}

// Clear the value at key from the cache, if it exists
func (l *UserLoaderByID) Clear(key persist.DBID) {
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *UserLoaderByID) unsafeSet(key persist.DBID, value coredb.User) {
	if l.cache == nil {
		l.cache = map[persist.DBID]coredb.User{}
	}
	l.cache[key] = value
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *userLoaderByIDBatch) keyIndex(l *UserLoaderByID, key persist.DBID) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	if pos == 0 {
		go b.startTimer(l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.batch = nil
			go b.end(l)
		}
	}

	return pos
}

func (b *userLoaderByIDBatch) startTimer(l *UserLoaderByID) {
	time.Sleep(l.wait)
	l.mu.Lock()

	// we must have hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.batch = nil
	l.mu.Unlock()

	b.end(l)
}

func (b *userLoaderByIDBatch) end(l *UserLoaderByID) {
	b.data, b.error = l.fetch(b.keys)
	close(b.done)
}
