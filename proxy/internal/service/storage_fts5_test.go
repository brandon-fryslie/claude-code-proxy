//go:build test
// +build test

package service

import (
	"database/sql"
	"log"
)

// createFTS5Table is a no-op in test builds (FTS5 not available)
func createFTS5Table(db *sql.DB) error {
	log.Println("⚠️  FTS5 disabled in test build - full-text search will use fallback")
	return nil
}

// fts5Enabled returns false in test builds
func fts5Enabled() bool {
	return false
}
