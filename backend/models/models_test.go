package models_test

import (
	"testing"

	"github.com/galpt/sotekre/backend/models"
	"github.com/stretchr/testify/require"
)

func ptrString(s string) *string { return &s }
func ptrUint(v uint) *uint       { return &v }

func TestMenu_ToNode(t *testing.T) {
	m := models.Menu{ID: 7, Title: "T", URL: ptrString("/x"), ParentID: ptrUint(3), Order: 5}
	n := m.ToNode()
	require.Equal(t, uint(7), n.ID)
	require.Equal(t, "T", n.Title)
	require.Equal(t, "/x", *n.URL)
	require.Equal(t, 5, n.Order)
	require.Nil(t, n.Children)
}
