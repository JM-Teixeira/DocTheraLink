package attachment

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"klyniva/internal/auth"
	"klyniva/internal/models"
	"klyniva/internal/modules/referral"

	"github.com/gin-gonic/gin"
)

// Handler exposes the Service over HTTP.
type Handler struct{ svc *Service }

// NewHandler builds a Handler around svc.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func actorFrom(c *gin.Context) (referral.Actor, bool) {
	claims, err := auth.GetCurrentUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "Invalid token"})
		return referral.Actor{}, false
	}
	return referral.Actor{ID: claims.UserID, Role: claims.UserType}, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, referral.ErrNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
	case errors.Is(err, ErrForbidden), errors.Is(err, referral.ErrForbidden):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: err.Error()})
	case errors.Is(err, ErrTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, models.ErrorResponse{Error: "too_large", Message: err.Error()})
	case errors.Is(err, ErrUnsupportedType):
		c.JSON(http.StatusUnsupportedMediaType, models.ErrorResponse{Error: "unsupported_type", Message: err.Error()})
	default:
		log.Printf("attachment: internal error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unexpected error"})
	}
}

// List godoc
// @Summary Files attached to a referral
// @Tags attachments
// @Security BearerAuth
// @Param id path string true "Referral ID"
// @Success 200 {array} Attachment
// @Router /referrals/{id}/attachments/ [get]
func (h *Handler) List(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	items, err := h.svc.List(c.Request.Context(), actor, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// Upload godoc
// @Summary Attach a file (multipart field "file") to a referral
// @Tags attachments
// @Security BearerAuth
// @Accept multipart/form-data
// @Param id path string true "Referral ID"
// @Param file formData file true "PDF, JPEG, PNG, WebP or text"
// @Success 201 {object} Attachment
// @Router /referrals/{id}/attachments/ [post]
func (h *Handler) Upload(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	// Cap the whole request a little above the file limit (multipart overhead).
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.svc.maxBytes+1<<20)
	fh, err := c.FormFile("file")
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(c, ErrTooLarge)
			return
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "multipart field \"file\" is required"})
		return
	}
	f, err := fh.Open()
	if err != nil {
		writeError(c, err)
		return
	}
	defer f.Close()
	a, err := h.svc.Upload(c.Request.Context(), actor, c.Param("id"), fh.Filename, f)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, a)
}

// Download godoc
// @Summary Download an attachment
// @Tags attachments
// @Security BearerAuth
// @Param id path string true "Attachment ID"
// @Success 200 {file} file
// @Router /attachments/{id}/download/ [get]
func (h *Handler) Download(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	a, body, err := h.svc.Open(c.Request.Context(), actor, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	defer body.Close()
	c.Header("Content-Type", a.ContentType)
	c.Header("Content-Length", strconv.FormatInt(a.Size, 10))
	// Always a download, never rendered inline by the browser
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(a.FileName))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "no-store")
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, body); err != nil {
		log.Printf("attachment: download %s interrupted: %v", a.ID, err)
	}
}

// Delete godoc
// @Summary Delete an attachment (uploader or admin)
// @Tags attachments
// @Security BearerAuth
// @Param id path string true "Attachment ID"
// @Success 204
// @Router /attachments/{id}/ [delete]
func (h *Handler) Delete(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), actor, c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
