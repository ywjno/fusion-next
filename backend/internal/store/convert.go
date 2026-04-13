package store

import "database/sql"

// boolToInt converts Go bool to SQLite INTEGER (0/1).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intToBool converts SQLite INTEGER (0/1) to Go bool.
func intToBool(i int) bool {
	return i != 0
}

// nullInt64ToBoolPtr converts a SQL NULL/INTEGER to a *bool.
func nullInt64ToBoolPtr(n sql.NullInt64) *bool {
	if !n.Valid {
		return nil
	}
	v := n.Int64 != 0
	return &v
}

// boolPtrToNullInt64 converts a *bool to a sql.NullInt64 suitable for SQLite.
func boolPtrToNullInt64(b *bool) sql.NullInt64 {
	if b == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(boolToInt(*b)), Valid: true}
}
