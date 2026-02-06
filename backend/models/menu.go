package models

import (
	"time"

	"gorm.io/gorm"
)

// Menu represents a hierarchical menu item stored in the DB.
type Menu struct {
	ID        uint           `gorm:":primaryKey" json:"id"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	URL       *string        `gorm:"size:1024" json:"url,omitempty"`
	Icon      *string        `gorm:"size:255" json:"icon,omitempty"`
	ParentID  *uint          `gorm:"index" json:"parent_id,omitempty"`
	Order     int            `gorm:"default:0;index" json:"order"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// MenuNode is the API representation with nested children.
type MenuNode struct {
	ID       uint        `json:"id"`
	Title    string      `json:"title"`
	URL      *string     `json:"url,omitempty"`
	ParentID *uint       `json:"parent_id,omitempty"`
	Order    int         `json:"order"`
	Children []*MenuNode `json:"children,omitempty"`
}

// ToNode converts Menu -> MenuNode (shallow)
func (m *Menu) ToNode() *MenuNode {
	return &MenuNode{
		ID:       m.ID,
		Title:    m.Title,
		URL:      m.URL,
		ParentID: m.ParentID,
		Order:    m.Order,
	}
}
