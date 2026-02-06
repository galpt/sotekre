package handlers

import (
	"net/http"
	"strconv"

	"github.com/galpt/sotekre/backend/models"
	"github.com/galpt/sotekre/backend/services"
	"github.com/gin-gonic/gin"
)

type createMenuInput struct {
	Title    string  `json:"title" binding:"required"`
	URL      *string `json:"url"`
	ParentID *uint   `json:"parent_id"`
	Order    *int    `json:"order"`
}

// --- types used only for API documentation (swag) ---
type getMenusResponse struct {
	Data []*models.MenuNode `json:"data"`
}

type updateMenuInput struct {
	Title    *string `json:"title,omitempty"`
	URL      *string `json:"url,omitempty"`
	ParentID *uint   `json:"parent_id,omitempty"`
	Order    *int    `json:"order,omitempty"`
}

type reorderInput struct {
	NewOrder int `json:"new_order" example:"0"`
}

type moveInput struct {
	NewParentID *uint `json:"new_parent_id"`
	NewOrder    *int  `json:"new_order,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Ensure these doc-only types are referenced so gopls / static analysis do not
// report them as unused (they're consumed by swag via reflection only).
var (
	_ = (*getMenusResponse)(nil)
	_ = (*updateMenuInput)(nil)
	_ = (*reorderInput)(nil)
	_ = (*moveInput)(nil)
	_ = (*errorResponse)(nil)
)

// -----------------------------------------------

// GetMenus godoc
// @Summary Get full menu tree
// @Tags menus
// @Produce json
// @Success 200 {object} getMenusResponse
// @Failure 500 {object} errorResponse
// @Router /api/menus [get]
func GetMenus(c *gin.Context) {
	flat, err := services.GetAllMenus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tree, _ := services.BuildTree(flat)
	if tree == nil {
		tree = []*models.MenuNode{}
	}
	c.JSON(http.StatusOK, gin.H{"data": tree})
}

// CreateMenu godoc
// @Summary Create a menu item
// @Tags menus
// @Accept json
// @Produce json
// @Param input body createMenuInput true "create menu"
// @Success 201 {object} models.Menu
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/menus [post]
func CreateMenu(c *gin.Context) {
	var in createMenuInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m := &models.Menu{
		Title: in.Title,
	}
	if in.URL != nil {
		m.URL = in.URL
	}
	if in.ParentID != nil {
		m.ParentID = in.ParentID
	}
	if in.Order != nil {
		m.Order = *in.Order
	}
	if err := services.CreateMenu(c.Request.Context(), m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": m})
}

// UpdateMenu godoc
// @Summary Update menu (partial)
// @Tags menus
// @Accept json
// @Produce json
// @Param id path int true "menu id"
// @Param input body updateMenuInput true "fields to update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/menus/{id} [put]
func UpdateMenu(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var in map[string]interface{}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// sanitize allowed fields
	allowed := map[string]bool{"title": true, "url": true, "parent_id": true, "order": true}
	upd := map[string]interface{}{}
	for k, v := range in {
		if allowed[k] {
			upd[k] = v
		}
	}
	if len(upd) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no updatable fields provided"})
		return
	}
	if err := services.UpdateMenu(c.Request.Context(), uint(id64), upd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// ReorderMenu godoc
// @Summary Reorder menu item within same parent
// @Tags menus
// @Accept json
// @Produce json
// @Param id path int true "menu id"
// @Param input body reorderInput true "new order"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/menus/{id}/reorder [patch]
func ReorderMenu(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var in struct {
		NewOrder *int `json:"new_order"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.NewOrder == nil || *in.NewOrder < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_order is required and must be >= 0"})
		return
	}
	if err := services.ReorderMenu(c.Request.Context(), uint(id64), *in.NewOrder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// MoveMenu godoc
// @Summary Move menu item to different parent and position
// @Tags menus
// @Accept json
// @Produce json
// @Param id path int true "menu id"
// @Param input body moveInput true "new parent and/or order"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/menus/{id}/move [patch]
func MoveMenu(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var in struct {
		NewParentID *uint `json:"new_parent_id"`
		NewOrder    *int  `json:"new_order"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.NewOrder != nil && *in.NewOrder < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_order must be >= 0"})
		return
	}
	if err := services.MoveMenu(c.Request.Context(), uint(id64), in.NewParentID, in.NewOrder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "moved"})
}

// DeleteMenu godoc
// @Summary Delete menu item (recursive)
// @Tags menus
// @Param id path int true "menu id"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/menus/{id} [delete]
func DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := services.DeleteMenuRecursive(c.Request.Context(), uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
