package ingest

import (
	"time"
)

// WatchDir polls dir and runs fn when entries may have changed.
// Returns only on error (callers interrupt via process signal in production).
func WatchDir(dir string, fn func() error) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
