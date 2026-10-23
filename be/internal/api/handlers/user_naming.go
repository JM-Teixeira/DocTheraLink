package handlers

import (
	"errors"
	"net/http"

	"klyniva/internal/auth"
	"klyniva/internal/database"
	"klyniva/internal/models"
	"klyniva/internal/usernames"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// errNoClinicCode: people of a clinic get its code in their username, so
// the clinic needs one first.
var errNoClinicCode = errors.New("the clinic has no code yet: set it on the clinic page first")

// newUsername picks the username for a person of clinic (nil: organisation
// admins) following the standard in internal/usernames.
func newUsername(db *gorm.DB, clinic *string, first, last string) (string, error) {
	code := ""
	if clinic != nil {
		var codes []string
		if err := db.Model(&models.Facility{}).Where("id = ?", *clinic).Pluck("code", &codes).Error; err != nil {
			return "", err
		}
		if len(codes) == 0 || codes[0] == "" {
			return "", errNoClinicCode
		}
		code = codes[0]
	}
	return usernames.Next(code, first, last, func(name string) (bool, error) {
		var n int64
		// Unscoped: names of deleted accounts stay reserved
		err := db.Unscoped().Model(&models.User{}).Where("username = ?", name).Count(&n).Error
		return n > 0, err
	})
}

// UsernameSuggestion godoc
// @Summary The username a new person would get (preview for the form)
// @Tags users
// @Security BearerAuth
// @Param first_name query string true "First name"
// @Param last_name query string true "Last name"
// @Param home_facility query string false "Clinic (admins; others get their own)"
// @Router /users/username_suggestion/ [get]
func (h *UserHandler) UsernameSuggestion(c *gin.Context) {
	db := database.GetDB()
	a, ok := currentActor(c, db)
	if !ok {
		return
	}
	if !auth.IsAdmin(a.Role) && a.Role != models.RoleCoordinator {
		forbid(c)
		return
	}
	clinic := a.Clinic
	if auth.IsAdmin(a.Role) {
		clinic = nil
		if id := c.Query("home_facility"); id != "" {
			clinic = &id
		}
	}
	name, err := newUsername(db, clinic, c.Query("first_name"), c.Query("last_name"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": name})
}
