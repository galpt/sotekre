package services

import (
	"context"
	"testing"

	"github.com/galpt/sotekre/backend/config"
	"github.com/galpt/sotekre/backend/models"
)

func TestBuildTree_missingParent_becomesRoot(t *testing.T) {
	flat := []models.Menu{
		{ID: 10, Title: "orphan", ParentID: ptrUint(999), Order: 0},
	}
	roots, err := BuildTree(flat)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}
	if len(roots) != 1 || roots[0].ID != 10 {
		t.Fatalf("expected orphan to become root, got %+v", roots)
	}
}

func TestCreateMenu_emptyTitle_errors(t *testing.T) {
	err := CreateMenu(context.Background(), &models.Menu{Title: ""})
	if err == nil {
		t.Fatalf("expected error when creating menu with empty title")
	}
}

func TestReorderMenu_nonExistentID_returnsError(t *testing.T) {
	setupInMemoryDBForServicesTest(t)
	defer config.CloseDB()
	if err := ReorderMenu(context.Background(), 99999, 1); err == nil {
		t.Fatalf("expected error when reordering non-existent id")
	}
}