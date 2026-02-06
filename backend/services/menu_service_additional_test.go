package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/galpt/sotekre/backend/config"
	"github.com/galpt/sotekre/backend/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

func TestUpdateMenu_noFields_returnsError(t *testing.T) {
	// no DB required — fast path
	if err := UpdateMenu(context.Background(), 1, map[string]interface{}{}); err == nil {
		t.Fatalf("expected error for empty update map")
	}
}

func TestUpdateMenu_updatesFields(t *testing.T) {
	setupInMemoryDBForServicesTest(t)
	defer config.CloseDB()
	m := models.Menu{Title: "old", Order: 2}
	if err := config.DB.Create(&m).Error; err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if err := UpdateMenu(context.Background(), m.ID, map[string]interface{}{"title": "new", "order": 5}); err != nil {
		t.Fatalf("UpdateMenu failed: %v", err)
	}
	var got models.Menu
	if err := config.DB.First(&got, m.ID).Error; err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if got.Title != "new" || got.Order != 5 {
		t.Fatalf("unexpected row after update: %+v", got)
	}
}

func TestDeleteMenuRecursive_deletesSubtree(t *testing.T) {
	setupInMemoryDBForServicesTest(t)
	defer config.CloseDB()
	p := models.Menu{Title: "p", Order: 0}
	if err := config.DB.Create(&p).Error; err != nil {
		t.Fatalf("create parent failed: %v", err)
	}
	c := models.Menu{Title: "c", ParentID: &p.ID, Order: 0}
	if err := config.DB.Create(&c).Error; err != nil {
		t.Fatalf("create child failed: %v", err)
	}
	g := models.Menu{Title: "g", ParentID: &c.ID, Order: 0}
	if err := config.DB.Create(&g).Error; err != nil {
		t.Fatalf("create grandchild failed: %v", err)
	}

	if err := DeleteMenuRecursive(context.Background(), p.ID); err != nil {
		t.Fatalf("DeleteMenuRecursive failed: %v", err)
	}
	var remaining []models.Menu
	if err := config.DB.Find(&remaining).Error; err != nil {
		t.Fatalf("read remaining failed: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected no rows after delete, got %d", len(remaining))
	}
}

func TestDeleteMenuRecursive_txRollback_onDeleteError_sqlmock(t *testing.T) {
	// use sqlmock to force the DELETE to fail and assert the transaction is rolled back
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open(sqlmock): %v", err)
	}
	config.DB = gdb

	mock.ExpectBegin()
	// the service will query children by parent_id (return empty rows here)
	mock.ExpectQuery("SELECT .*FROM .*menus.*parent_id").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "parent_id", "order"}))
	// force the soft-delete (UPDATE ... SET `deleted_at`) to fail
	mock.ExpectExec("UPDATE .*menus.*").WillReturnError(fmt.Errorf("boom"))
	mock.ExpectRollback()

	err = DeleteMenuRecursive(context.Background(), 42)
	if err == nil {
		t.Fatalf("expected error from DeleteMenuRecursive when DELETE fails")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMoveMenu_cannotMoveIntoOwnDescendant(t *testing.T) {
	setupInMemoryDBForServicesTest(t)
	defer config.CloseDB()
	// create a -> b -> c
	if err := config.DB.Create(&models.Menu{Title: "a"}).Error; err != nil {
		t.Fatalf("create a: %v", err)
	}
	var a models.Menu
	config.DB.First(&a, "title = ?", "a")
	if err := config.DB.Create(&models.Menu{Title: "b", ParentID: &a.ID}).Error; err != nil {
		t.Fatalf("create b: %v", err)
	}
	var b models.Menu
	config.DB.First(&b, "title = ?", "b")
	if err := config.DB.Create(&models.Menu{Title: "c", ParentID: &b.ID}).Error; err != nil {
		t.Fatalf("create c: %v", err)
	}
	var c models.Menu
	config.DB.First(&c, "title = ?", "c")

	err := MoveMenu(context.Background(), a.ID, &c.ID, nil)
	if err == nil {
		t.Fatalf("expected error when moving parent into descendant")
	}
}

func TestMoveMenu_sameParent_noop(t *testing.T) {
	setupInMemoryDBForServicesTest(t)
	defer config.CloseDB()
	// create roots A,B,C
	if err := config.DB.Create(&models.Menu{Title: "A", Order: 0}).Error; err != nil {
		t.Fatalf("create A: %v", err)
	}
	if err := config.DB.Create(&models.Menu{Title: "B", Order: 1}).Error; err != nil {
		t.Fatalf("create B: %v", err)
	}
	if err := config.DB.Create(&models.Menu{Title: "C", Order: 2}).Error; err != nil {
		t.Fatalf("create C: %v", err)
	}
	var a, b models.Menu
	if err := config.DB.First(&a, "title = ?", "A").Error; err != nil {
		t.Fatalf("read A failed: %v", err)
	}
	if err := config.DB.First(&b, "title = ?", "B").Error; err != nil {
		t.Fatalf("read B failed: %v", err)
	}
	idB := b.ID

	// find current index of B among roots
	var roots []models.Menu
	config.DB.Where("parent_id IS NULL").Order("\"order\" asc").Find(&roots)
	curIdx := -1
	for i, r := range roots {
		if r.ID == idB {
			curIdx = i
			break
		}
	}
	if curIdx == -1 {
		t.Fatalf("could not find B in roots; roots=%+v", roots)
	}

	if err := MoveMenu(context.Background(), idB, nil, &curIdx); err != nil {
		t.Fatalf("MoveMenu same-parent noop failed: %v", err)
	}
	// verify order unchanged
	var after []models.Menu
	config.DB.Where("parent_id IS NULL").Order("\"order\" asc").Find(&after)
	if len(after) != 3 || after[curIdx].ID != idB {
		t.Fatalf("expected B to remain at index %d, got %+v", curIdx, after)
	}
}

func TestMoveMenu_newParentNotFound_setsParent(t *testing.T) {
	setupInMemoryDBForServicesTest(t)
	defer config.CloseDB()
	m := models.Menu{Title: "loner"}
	if err := config.DB.Create(&m).Error; err != nil {
		t.Fatalf("create loner: %v", err)
	}
	nonEx := uint(99999)
	if err := MoveMenu(context.Background(), m.ID, &nonEx, nil); err != nil {
		t.Fatalf("expected MoveMenu to allow non-existent parent (current behavior): %v", err)
	}
	var got models.Menu
	if err := config.DB.First(&got, m.ID).Error; err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if got.ParentID == nil || *got.ParentID != nonEx {
		t.Fatalf("expected parent_id to be set to %d, got %+v", nonEx, got.ParentID)
	}
}

func TestMoveMenu_insertIndexBounds(t *testing.T) {
	setupInMemoryDBForServicesTest(t)
	defer config.CloseDB()
	// create roots A,B,C
	config.DB.Create(&models.Menu{Title: "A", Order: 0})
	config.DB.Create(&models.Menu{Title: "B", Order: 1})
	config.DB.Create(&models.Menu{Title: "C", Order: 2})
	var a, b, c models.Menu
	config.DB.First(&a, "title = ?", "A")
	config.DB.First(&b, "title = ?", "B")
	config.DB.First(&c, "title = ?", "C")

	// move C to newOrder -1 => should clamp to 0
	neg := -1
	require.NoError(t, MoveMenu(context.Background(), c.ID, nil, &neg))
	var roots []models.Menu
	config.DB.Where("parent_id IS NULL").Order("\"order\" asc").Find(&roots)
	require.Equal(t, 3, len(roots))
	require.Equal(t, c.ID, roots[0].ID)

	// move C to very large index -> append
	big := 100
	require.NoError(t, MoveMenu(context.Background(), c.ID, nil, &big))
	config.DB.Where("parent_id IS NULL").Order("\"order\" asc").Find(&roots)
	require.Equal(t, c.ID, roots[2].ID)
}

func TestMoveMenu_betweenParents_compactSourceOrders(t *testing.T) {
	setupInMemoryDBForServicesTest(t)
	defer config.CloseDB()
	// source parent P with children s0,s1,s2
	p1 := models.Menu{Title: "P1"}
	config.DB.Create(&p1)
	config.DB.Create(&models.Menu{Title: "s0", ParentID: &p1.ID, Order: 0})
	config.DB.Create(&models.Menu{Title: "s1", ParentID: &p1.ID, Order: 1})
	config.DB.Create(&models.Menu{Title: "s2", ParentID: &p1.ID, Order: 2})
	// dest parent Q with one child
	p2 := models.Menu{Title: "P2"}
	config.DB.Create(&p2)
	config.DB.Create(&models.Menu{Title: "d0", ParentID: &p2.ID, Order: 0})

	// move s1 -> p2 at index 1
	var s1 models.Menu
	config.DB.First(&s1, "title = ?", "s1")
	idx := 1
	require.NoError(t, MoveMenu(context.Background(), s1.ID, &p2.ID, &idx))

	// verify source compacted (s0,s2) = orders 0,1
	var src []models.Menu
	config.DB.Where("parent_id = ?", p1.ID).Order("\"order\" asc").Find(&src)
	require.Len(t, src, 2)
	require.Equal(t, 0, src[0].Order)
	require.Equal(t, 1, src[1].Order)

	// verify dest has d0 at 0 and moved at 1
	var dst []models.Menu
	config.DB.Where("parent_id = ?", p2.ID).Order("\"order\" asc").Find(&dst)
	require.Len(t, dst, 2)
	require.Equal(t, 0, dst[0].Order)
	require.Equal(t, 1, dst[1].Order)
}

func TestBuildTree_deepSorting_and_recursion(t *testing.T) {
	flat := []models.Menu{
		{ID: 1, Title: "root", ParentID: nil, Order: 2},
		{ID: 2, Title: "a", ParentID: ptrUint(1), Order: 1},
		{ID: 3, Title: "b", ParentID: ptrUint(1), Order: 0},
		{ID: 4, Title: "b-child", ParentID: ptrUint(3), Order: 0},
	}
	roots, err := BuildTree(flat)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	// root should have children ordered by Order asc -> [b(id=3), a(id=2)]
	if roots[0].Children[0].ID != 3 || roots[0].Children[1].ID != 2 {
		t.Fatalf("unexpected child order: %+v", roots[0].Children)
	}
	// deep child exists
	if len(roots[0].Children[0].Children) != 1 || roots[0].Children[0].Children[0].ID != 4 {
		t.Fatalf("expected deep child b-child")
	}
}

func TestGetAllMenus_DBError_sqlmock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()
	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	require.NoError(t, err)
	config.DB = gdb

	mock.ExpectQuery("SELECT .*FROM .*menus").WillReturnError(fmt.Errorf("boom"))
	_, err = GetAllMenus(context.Background())
	if err == nil {
		t.Fatalf("expected error from GetAllMenus when SELECT fails")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}
func TestMoveMenu_txRollback_onExecError_sqlmock(t *testing.T) {
	// use sqlmock to force an Exec error while reordering destination
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open(sqlmock): %v", err)
	}
	config.DB = gdb

	// transaction begins inside the service — expect it first
	mock.ExpectBegin()
	// load the item (first query) — permissive regexp to match prepared SQL
	mock.ExpectQuery("SELECT .*").WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "order"}).AddRow(1, nil, 0))
	// fetch dest siblings — include two siblings so final insertion index differs from current index
	mock.ExpectQuery("SELECT .*menus.*parent_id").WillReturnRows(sqlmock.NewRows([]string{"id", "order"}).AddRow(2, 0).AddRow(3, 1))
	// same-parent path will re-query the siblings for source list — return item + siblings
	mock.ExpectQuery("SELECT .*menus.*parent_id").WillReturnRows(sqlmock.NewRows([]string{"id", "order"}).AddRow(1, 0).AddRow(2, 0).AddRow(3, 1))
	// force the Exec that writes back ordering to fail
	mock.ExpectExec("UPDATE .*menus.*").WillReturnError(fmt.Errorf("boom"))
	mock.ExpectRollback()

	err = MoveMenu(context.Background(), 1, nil, nil)
	if err == nil {
		t.Fatalf("expected error from MoveMenu when UPDATE fails")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
