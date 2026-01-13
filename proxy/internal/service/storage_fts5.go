//go:build !test
// +build !test

package service

import (
	"database/sql"
	"fmt"
	"log"
)

// createFTS5Table creates the FTS5 virtual table for full-text search (production builds only)
func createFTS5Table(db *sql.DB) error {
	// Check if FTS5 table exists
	var ftsExists int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='conversations_fts'").Scan(&ftsExists)
	if err != nil {
		return fmt.Errorf("failed to check if FTS table exists: %w", err)
	}

	if ftsExists == 0 {
		// Create FTS5 virtual table
		ftsSchema := `
		CREATE VIRTUAL TABLE conversations_fts USING fts5(
			conversation_id UNINDEXED,
			message_uuid UNINDEXED,
			message_type,
			content_text,
			tool_names,
			timestamp UNINDEXED,
			tokenize='porter unicode61'
		);
		`

		if _, err := db.Exec(ftsSchema); err != nil {
			return fmt.Errorf("failed to create FTS table: %w", err)
		}

		log.Println("✅ Created conversations_fts FTS5 table")
	}

	return nil
}

// fts5Enabled returns true in production builds
func fts5Enabled() bool {
	return true
}
