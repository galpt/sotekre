package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/galpt/sotekre/backend/config"
	"github.com/galpt/sotekre/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBuildTree_simple(t *testing.T) {
	flat := []models.Menu{
		{ID: 1, Title: "Root A", Order: 1},
		{ID: 2, Title: "Child A1", ParentID: ptrUint(1), Order: 1},
		{ID: 3, Title: "Child A2", ParentID: ptrUint(1), Order: 2},
		{ID: 4, Title: "Root B", Order: 2},
	}

	roots, err := BuildTree(flat)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	if roots[0].ID != 1 || len(roots[0].Children) != 2 {
		t.Fatalf("root A children mismatch: %+v", roots[0])
	}
	if roots[1].ID != 4 {
		t.Fatalf("root B mismatch: %+v", roots[1])
	}
}

func ptrUint(v uint) *uint { return &v }

func setupInMemoryDBForServicesTest(t *testing.T) {
	// use a unique in-memory DSN per test to avoid cross-test pollution
	dsn := fmt.Sprintf("file:memtest_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	config.DB = db
	if err := config.DB.AutoMigrate(&models.Menu{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
}

func ptrInt(v int) *int { return &v }

func TestReorderMenu_service(t *testing.T) {
	setupInMemoryDBForServicesTest(t)
	// create three roots A(0), B(1), C(2)
	a := models.Menu{Title: "A", Order: 0}
	b := models.Menu{Title: "B", Order: 1}
	c := models.Menu{Title: "C", Order: 2}
	if err := config.DB.Create(&a).Error; err != nil { t.Fatalf("create A: %v", err) }
	if err := config.DB.Create(&b).Error; err != nil { t.Fatalf("create B: %v", err) }
	if err := config.DB.Create(&c).Error; err != nil { t.Fatalf("create C: %v", err) }

	// move C to index 1 -> expected order A, C, B
	if err := ReorderMenu(context.Background(), c.ID, 1); err != nil {
		t.Fatalf("ReorderMenu failed: %v", err)
	}

	flat, err := GetAllMenus(context.Background())
	if err != nil { t.Fatalf("GetAllMenus: %v", err) }
	roots, _ := BuildTree(flat)
	if len(roots) != 3 { t.Fatalf("expected 3 roots, got %d", len(roots)) }
	if roots[0].ID != a.ID || roots[1].ID != c.ID || roots[2].ID != b.ID {
		t.Fatalf("unexpected order: %v", []uint{roots[0].ID, roots[1].ID, roots[2].ID})
	}
}

func TestMoveMenu_service_between_parents(t *testing.T) {
	setupInMemoryDBForServicesTest(t)
	// roots: A, B, C
	a := models.Menu{Title: "A", Order: 0}
	b := models.Menu{Title: "B", Order: 1}
	c := models.Menu{Title: "C", Order: 2}
	config.DB.Create(&a)
	config.DB.Create(&b)
	config.DB.Create(&c)

	// move B to be the first child of A
	if err := MoveMenu(context.Background(), b.ID, &a.ID, ptrInt(0)); err != nil {
		t.Fatalf("MoveMenu failed: %v", err)
	}

	// debug: dump DB rows
	flatRaw := []models.Menu{}
	if err := config.DB.Order("\"order\" asc").Find(&flatRaw).Error; err != nil {
		t.Fatalf("read raw menus: %v", err)
	}
	for _, r := range flatRaw {
		t.Logf("row id=%d parent=%v order=%d", r.ID, r.ParentID, r.Order)
	}

	flat, err := GetAllMenus(context.Background())
	if err != nil { t.Fatalf("GetAllMenus: %v", err) }
	roots, _ := BuildTree(flat)
	// roots should be A and C (B moved under A)
	if len(roots) != 2 { t.Fatalf("expected 2 roots after move, got %d", len(roots)) }
	if roots[0].ID != a.ID || roots[1].ID != c.ID {
		t.Fatalf("unexpected roots after move: %v", []uint{roots[0].ID, roots[1].ID})
	}
	if len(roots[0].Children) != 1 || roots[0].Children[0].ID != b.ID {
		t.Fatalf("expected B as child of A, got %+v", roots[0].Children)
	}

	// cannot move A under its descendant B (should error)
	if err := MoveMenu(context.Background(), a.ID, &b.ID, ptrInt(0)); err == nil {
		t.Fatalf("expected error when moving parent under descendant, got nil")
	}
}


