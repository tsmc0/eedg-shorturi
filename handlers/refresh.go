package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"eedg-shortlink/helpers"
	"eedg-shortlink/models"
)

// HandleRefreshCache обрабатывает POST /api/links/{code}/refresh.
func (a *App) HandleRefreshCache(w http.ResponseWriter, r *http.Request) {
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

	res, err := helpers.FetchContent(row.LongURL, a.FetchTimeout, a.maxContentBytes())
	if err != nil {
		helpers.WriteError(w, http.StatusBadGateway, "fetch failed: "+err.Error())
		return
	}

	expires := time.Now().UTC().Add(a.DefaultTTL)
	if err := a.Store.UpdateCache(code, res.Gz, res.Mime, expires); err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	helpers.WriteJSON(w, http.StatusOK, models.RefreshResponse{
		Code:      code,
		ExpiresAt: expires,
		Mime:      res.Mime,
		SizeGz:    len(res.Gz),
	})
}
