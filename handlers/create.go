package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"eedg-shortlink/helpers"
	"eedg-shortlink/models"
)

// handleCreate обрабатывает POST /api/links.
func (a *App) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		helpers.WriteError(w, http.StatusBadRequest, "url is required")
		return
	}
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		helpers.WriteError(w, http.StatusBadRequest, "url must start with http:// or https://")
		return
	}

	ttl := a.DefaultTTL
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}

	// Опционально кешируем контент до создания записи.
	var (
		cache     []byte
		mime      string
		expiresAt time.Time
	)
	if req.CacheContent {
		res, err := helpers.FetchContent(req.URL, a.FetchTimeout, a.maxContentBytes())
		if err != nil {
			helpers.WriteError(w, http.StatusBadGateway, "cache fetch failed: "+err.Error())
			return
		}
		cache = res.Gz
		mime = res.Mime
		expiresAt = time.Now().UTC().Add(ttl)
	}

	//

	row, err := a.Store.InsertUnique(
		func() (string, error) { return helpers.NewCode(6) },
		func(code string) models.LinkRow {
			return models.LinkRow{
				Code:       code,
				LongURL:    req.URL,
				ForceCheck: req.ForceCheck,
				Cache:      cache,
				MimeType:   mime,
				ExpiresAt:  expiresAt,
				CreatedAt:  time.Now().UTC(),
			}
		},
		5,
	)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, models.CreateResponse{
		Code:     row.Code,
		ShortURL: strings.TrimRight(a.BaseShortURL, "/") + "/u/" + row.Code,
		LongURL:  row.LongURL,
	})
}
