package repo

import (
	"strings"
	"time"

	"github.com/0x2e/fusion/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewItem(db *gorm.DB) *Item {
	return &Item{
		db: db,
	}
}

type Item struct {
	db *gorm.DB
}

type ItemFilter struct {
	Keyword  *string
	FeedID   *uint
	GroupID  *uint
	Unread   *bool
	Bookmark *bool
}

func (i Item) List(filter ItemFilter, page, pageSize int) ([]*model.Item, int, error) {
	var total int64
	var res []*model.Item
	db := i.db.Model(&model.Item{}).Joins("JOIN feeds ON feeds.id = items.feed_id")
	if filter.Keyword != nil {
		expr := "%" + *filter.Keyword + "%"
		db = db.Where("title LIKE ? OR content LIKE ?", expr, expr)
	}
	if filter.FeedID != nil {
		db = db.Where("feed_id = ?", *filter.FeedID)
	}
	if filter.GroupID != nil {
		db = db.Where("feeds.group_id = ?", *filter.GroupID)
	}
	if filter.Unread != nil {
		db = db.Where("unread = ?", *filter.Unread)
	}
	if filter.Bookmark != nil {
		db = db.Where("bookmark = ?", *filter.Bookmark)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Preload("Feed").Order("items.pub_date desc, items.created_at desc").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&res).Error
	return res, int(total), err
}

func (i Item) Get(id uint) (*model.Item, error) {
	var res model.Item
	err := i.db.Joins("Feed").First(&res, id).Error
	return &res, err
}

func (i Item) Insert(items []*model.Item) error {
	// limit batchSize to fix 'too many SQL variable' error
	now := time.Now()
	for _, i := range items {
		i.CreatedAt = now
		i.UpdatedAt = now
	}
	return i.db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).CreateInBatches(items, 5).Error
}

func (i Item) Update(id uint, item *model.Item) error {
	return i.db.Model(&model.Item{}).Where("id = ?", id).Updates(item).Error
}

func (i Item) Delete(id uint) error {
	return i.db.Delete(&model.Item{}, id).Error
}

func (i Item) UpdateUnread(ids []uint, unread *bool) error {
	return i.db.Model(&model.Item{}).Where("id IN ?", ids).Update("unread", unread).Error
}

func (i Item) UpdateBookmark(id uint, bookmark *bool) error {
	return i.db.Model(&model.Item{}).Where("id = ?", id).Update("bookmark", bookmark).Error
}

func (i Item) UpdateFullContent(id uint, fullContent *string) error {
	return i.db.Model(&model.Item{}).Where("id = ?", id).Update("full_content", fullContent).Error
}

func (i Item) BatchUpdateFullContent(updates map[uint]string) error {
	if len(updates) == 0 {
		return nil
	}

	// For small batches, use individual updates in a transaction
	if len(updates) <= 5 {
		return i.db.Transaction(func(tx *gorm.DB) error {
			for id, content := range updates {
				if err := tx.Model(&model.Item{}).Where("id = ?", id).Update("full_content", content).Error; err != nil {
					return err
				}
			}
			return nil
		})
	}

	// For larger batches, use CASE WHEN for better performance
	return i.db.Transaction(func(tx *gorm.DB) error {
		// Build CASE WHEN statement
		var caseClauses []string
		var args []interface{}
		var ids []uint

		for id, content := range updates {
			caseClauses = append(caseClauses, "WHEN ? THEN ?")
			args = append(args, id, content)
			ids = append(ids, id)
		}

		// Build placeholders for IN clause
		placeholders := make([]string, len(ids))
		for idx := range ids {
			placeholders[idx] = "?"
		}

		// Build the final SQL
		sql := "UPDATE items SET full_content = CASE id " +
			strings.Join(caseClauses, " ") +
			" END, updated_at = ? WHERE id IN (" +
			strings.Join(placeholders, ",") + ")"

		args = append(args, time.Now())
		for _, id := range ids {
			args = append(args, id)
		}

		return tx.Exec(sql, args...).Error
	})
}
