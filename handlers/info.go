package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"eedg-shortlink/helpers"
	"eedg-shortlink/models"
)

// HandleInfo обрабатывает GET /api/links/{code}.
func (a *App) HandleInfo(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	row, err := a.Store.Get(code)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}
	if row == nil {
		helpers.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	helpers.WriteJSON(w, http.StatusOK, models.InfoResponse{
		Code:       row.Code,
		LongURL:    row.LongURL,
		ForceCheck: row.ForceCheck,
		HasCache:   len(row.Cache) > 0,
		CacheSize:  len(row.Cache),
		Mime:       row.MimeType,
		ExpiresAt:  row.ExpiresAt,
		CreatedAt:  row.CreatedAt,
	})
}
