package store

import (
	"database/sql"
	"time"
)

// FullContentBackfillCandidate is an item eligible for full-content backfill.
type FullContentBackfillCandidate struct {
	ID   int64
	Link string
}

// ListFullContentBackfillCandidates returns items that have a link but no
// full content yet. days > 0 restricts to items published within the last
// days; feedID > 0 restricts to one feed; limit > 0 caps the result count.
func (s *Store) ListFullContentBackfillCandidates(days int, limit int, feedID int64) ([]FullContentBackfillCandidate, error) {
	query := `
		SELECT id, link
		FROM items
		WHERE link != '' AND COALESCE(full_content, '') = ''
	`
	args := []any{}

	if days > 0 {
		query += ` AND pub_date > :cutoff`
		args = append(args, sql.Named("cutoff", time.Now().Unix()-int64(days)*86400))
	}
	if feedID > 0 {
		query += ` AND feed_id = :feed_id`
		args = append(args, sql.Named("feed_id", feedID))
	}

	query += ` ORDER BY pub_date DESC`

	if limit > 0 {
		query += ` LIMIT :limit`
		args = append(args, sql.Named("limit", limit))
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := []FullContentBackfillCandidate{}
	for rows.Next() {
		var c FullContentBackfillCandidate
		if err := rows.Scan(&c.ID, &c.Link); err != nil {
			return nil, err
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}
