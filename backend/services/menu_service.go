package services

import (
	"context"
	"errors"
	"sort"

	"github.com/galpt/sotekre/backend/config"
	"github.com/galpt/sotekre/backend/models"
	"gorm.io/gorm"
)

// GetAllMenus returns all menus ordered by `order` ASC.
func GetAllMenus(ctx context.Context) ([]models.Menu, error) {
	var menus []models.Menu
	if err := config.DB.Order("\"order\" asc").Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

// BuildTree converts a flat list of Menu into a nested slice (roots only).
func BuildTree(flat []models.Menu) ([]*models.MenuNode, error) {
	// map id -> node
	nodes := make(map[uint]*models.MenuNode, len(flat))
	for i := range flat {
		m := flat[i] // copy
		n := m.ToNode()
		nodes[m.ID] = n
	}

	var roots []*models.MenuNode
	for i := range flat {
		m := flat[i]
		n := nodes[m.ID]
		if m.ParentID != nil {
			p, ok := nodes[*m.ParentID]
			if ok {
				p.Children = append(p.Children, n)
				continue
			}
			// parent not found in list -> treat as root
		}
		roots = append(roots, n)
	}

	// sort helper
	var sortRec func(list []*models.MenuNode)
	sortRec = func(list []*models.MenuNode) {
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].Order < list[j].Order
		})
		for _, it := range list {
			if len(it.Children) > 0 {
				sortRec(it.Children)
			}
		}
	}

	sortRec(roots)
	return roots, nil
}

// CreateMenu inserts a new Menu row.
func CreateMenu(ctx context.Context, m *models.Menu) error {
	if m.Title == "" {
		return errors.New("title is required")
	}
	return config.DB.Create(m).Error
}

// UpdateMenu updates allowed fields for a menu item.
func UpdateMenu(ctx context.Context, id uint, upd map[string]interface{}) error {
	if len(upd) == 0 {
		return errors.New("no fields to update")
	}
	return config.DB.Model(&models.Menu{}).Where("id = ?", id).Updates(upd).Error
}

// DeleteMenuRecursive deletes a menu and all its children (transactional).
func DeleteMenuRecursive(ctx context.Context, id uint) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// NOTE: using raw SQL inside the transaction for simplicity
		// find children recursively and delete. Simpler approach: repeated queries.
		var toDelete []uint
		var stack = []uint{id}
		for len(stack) > 0 {
			cur := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			toDelete = append(toDelete, cur)
			var children []models.Menu
			if err := tx.Where("parent_id = ?", cur).Find(&children).Error; err != nil {
				return err
			}
			for _, ch := range children {
				stack = append(stack, ch.ID)
			}
		}
		// delete all collected ids
		if err := tx.Where("id IN (?)", toDelete).Delete(&models.Menu{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// ReorderMenu reorders an item within its current parent to the specified index.
func ReorderMenu(ctx context.Context, id uint, newOrder int) error {
	// fetch item's current parent and delegate to MoveMenu
	var item models.Menu
	if err := config.DB.First(&item, id).Error; err != nil {
		return err
	}
	return MoveMenu(ctx, id, item.ParentID, &newOrder)
}

// MoveMenu moves an item to a (possibly different) parent and inserts it at newOrder.
// If newOrder is nil the item will be appended to the destination's children.
func MoveMenu(ctx context.Context, id uint, newParentID *uint, newOrder *int) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// load the item
		var item models.Menu
		if err := tx.Clauses().First(&item, id).Error; err != nil {
			return err
		}

		oldParent := item.ParentID

		// prevent moving item into its own descendant (walk up from destination)
		if newParentID != nil {
			cur := newParentID
			for cur != nil {
				if *cur == id {
					return errors.New("cannot move item into its own descendant")
				}
				var p models.Menu
				if err := tx.Select("parent_id").Where("id = ?", *cur).First(&p).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						break
					}
					return err
				}
				cur = p.ParentID
			}
		}

		// fetch destination siblings (excluding the item)
		var destSibs []models.Menu
		q := tx.Model(&models.Menu{})
		if newParentID == nil {
			q = q.Where("parent_id IS NULL")
		} else {
			q = q.Where("parent_id = ?", *newParentID)
		}
		if err := q.Order("\"order\" asc").Find(&destSibs).Error; err != nil {
			return err
		}
		// filter out the moving item if present (same-parent move)
		tabledest := make([]models.Menu, 0, len(destSibs))
		for _, s := range destSibs {
			if s.ID == id {
				continue
			}
			tabledest = append(tabledest, s)
		}
		destSibs = tabledest

		// determine insertion index
		insertIdx := len(destSibs) // append by default
		if newOrder != nil {
			if *newOrder < 0 {
				insertIdx = 0
			} else if *newOrder > len(destSibs) {
				insertIdx = len(destSibs)
			} else {
				insertIdx = *newOrder
			}
		}

		// if moving within same parent and position unchanged -> no-op
		if (oldParent == nil && newParentID == nil) || (oldParent != nil && newParentID != nil && *oldParent == *newParentID) {
			// same parent: check index
			// build current order slice (excluding item)
			var srcSibs []models.Menu
			srcQ := tx.Model(&models.Menu{})
			if oldParent == nil {
				srcQ = srcQ.Where("parent_id IS NULL")
			} else {
				srcQ = srcQ.Where("parent_id = ?", *oldParent)
			}
			if err := srcQ.Order("\"order\" asc").Find(&srcSibs).Error; err != nil {
				return err
			}
			// find current index of item among siblings
			curIdx := -1
			for i, s := range srcSibs {
				if s.ID == id {
					curIdx = i
					break
				}
			}
			if curIdx == -1 {
				// item might be missing from list (shouldn't happen) — continue to generic path
			} else {
				// compute target index after removing the item
				if newParentID == nil && oldParent == nil || (oldParent != nil && newParentID != nil && *oldParent == *newParentID) {
					// remove current
					if insertIdx > curIdx {
						insertIdx-- // account for removal earlier in the list
					}
					if insertIdx == curIdx {
						return nil // nothing to do
					}
				}
			}
		}

		// Build final destination ID order (slice of IDs) by inserting item ID at insertIdx
		finalIDs := make([]uint, 0, len(destSibs)+1)
		for i, s := range destSibs {
			if i == insertIdx {
				finalIDs = append(finalIDs, id)
			}
			finalIDs = append(finalIDs, s.ID)
		}
		if insertIdx == len(destSibs) {
			finalIDs = append(finalIDs, id)
		}

		// If moving between different parents, compact the source parent's orders (remove the item)
		if !(oldParent == nil && newParentID == nil) {
			sameParent := oldParent != nil && newParentID != nil && *oldParent == *newParentID
			if !sameParent {
				var srcRem []models.Menu
				srcQ := tx.Model(&models.Menu{})
				if oldParent == nil {
					srcQ = srcQ.Where("parent_id IS NULL")
				} else {
					srcQ = srcQ.Where("parent_id = ?", *oldParent)
				}
				srcQ.Order("\"order\" asc").Find(&srcRem)
				// renumber srcRem excluding item
				idx := 0
				for _, s := range srcRem {
					if s.ID == id {
						continue
					}
					if s.Order != idx {
						if err := tx.Model(&models.Menu{}).Where("id = ?", s.ID).Update("order", idx).Error; err != nil {
							return err
						}
					}
					idx++
				}
			}
		}

		// write back destination ordering and update parent for the moved item
		for idx, idv := range finalIDs {
			upd := map[string]interface{}{"order": idx}
			// for the moved item, ensure parent_id is set to newParentID
			if idv == id {
				upd["parent_id"] = newParentID
			}
			if err := tx.Model(&models.Menu{}).Where("id = ?", idv).Updates(upd).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
