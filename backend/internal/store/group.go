package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/0x2E/fusion/internal/model"
)

func (s *Store) ListGroups() ([]*model.Group, error) {
	rows, err := s.db.Query(`
		SELECT id, name, auto_fetch_full_content, created_at, updated_at
		FROM groups
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []*model.Group{}
	for rows.Next() {
		g := &model.Group{}
		var autoFetchFullContent sql.NullInt64
		if err := rows.Scan(&g.ID, &g.Name, &autoFetchFullContent, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		g.AutoFetchFullContent = nullInt64ToBoolPtr(autoFetchFullContent)
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (s *Store) GetGroup(id int64) (*model.Group, error) {
	g := &model.Group{}
	var autoFetchFullContent sql.NullInt64
	err := s.db.QueryRow(`
		SELECT id, name, auto_fetch_full_content, created_at, updated_at
		FROM groups
		WHERE id = :id
	`, sql.Named("id", id)).Scan(
		&g.ID, &g.Name, &autoFetchFullContent, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: group", ErrNotFound)
		}
		return nil, fmt.Errorf("get group: %w", err)
	}
	g.AutoFetchFullContent = nullInt64ToBoolPtr(autoFetchFullContent)
	return g, nil
}

func (s *Store) CreateGroup(name string, autoFetch ...*bool) (*model.Group, error) {
	var af *bool
	if len(autoFetch) > 0 {
		af = autoFetch[0]
	}

	var result sql.Result
	var err error
	if af != nil {
		result, err = s.db.Exec(`
			INSERT INTO groups (name, auto_fetch_full_content) VALUES (:name, :auto_fetch_full_content)
		`, sql.Named("name", name), sql.Named("auto_fetch_full_content", boolToInt(*af)))
	} else {
		result, err = s.db.Exec(`
			INSERT INTO groups (name) VALUES (:name)
		`, sql.Named("name", name))
	}
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetGroup(id)
}

type UpdateGroupParams struct {
	Name                 *string
	AutoFetchFullContent *bool
}

func (s *Store) UpdateGroup(id int64, params UpdateGroupParams) error {
	setClauses := []string{}
	args := []any{sql.Named("id", id)}

	if params.Name != nil {
		setClauses = append(setClauses, "name = :name")
		args = append(args, sql.Named("name", *params.Name))
	}
	if params.AutoFetchFullContent != nil {
		setClauses = append(setClauses, "auto_fetch_full_content = :auto_fetch_full_content")
		args = append(args, sql.Named("auto_fetch_full_content", boolToInt(*params.AutoFetchFullContent)))
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = unixepoch()")
	query := fmt.Sprintf("UPDATE groups SET %s WHERE id = :id", strings.Join(setClauses, ", "))
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: group", ErrNotFound)
	}
	return nil
}

// DeleteGroup removes a group and moves all its feeds to the default group (ID=1).
// The default group itself cannot be deleted to ensure all feeds have a valid group.
func (s *Store) DeleteGroup(id int64) error {
	if id == 1 {
		return fmt.Errorf("%w: cannot delete default group", ErrInvalid)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE feeds SET group_id = 1 WHERE group_id = :id`, sql.Named("id", id)); err != nil {
		return err
	}

	result, err := tx.Exec(`DELETE FROM groups WHERE id = :id`, sql.Named("id", id))
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: group", ErrNotFound)
	}

	return tx.Commit()
}
