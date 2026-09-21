package handlers

import (
	"net/http"
	"time"

	"eedg-shortlink/helpers"

	"github.com/go-chi/chi/v5"
)

func (a *App) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// HandleRedirect обрабатывает GET /{code}: отдаёт кеш или редиректит.
func (a *App) HandleRedirect(w http.ResponseWriter, r *http.Request) {
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

	applyCacheHeaders(w, row.ForceCheck)
	w.Header().Set("X-Original-URL", row.LongURL)

	fresh := len(row.Cache) > 0 &&
		(row.ExpiresAt.IsZero() || row.ExpiresAt.After(time.Now().UTC()))

	if fresh {
		raw, err := helpers.GunzipBytes(row.Cache)
		if err != nil {
			_ = a.Store.ClearCache(code)
		} else {
			if row.MimeType != "" {
				w.Header().Set("Content-Type", row.MimeType)
			}
			w.Header().Set("X-Cache", "HIT")

			// Если клиент умеет gzip — не разжимаем обратно.
			if acceptsGzip(r) {
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Del("Content-Length")
				_, _ = w.Write(row.Cache)
				return
			}
			_, _ = w.Write(raw)
			return
		}
	}

	w.Header().Set("X-Cache", "MISS")
	http.Redirect(w, r, row.LongURL, http.StatusFound)
}

func applyCacheHeaders(w http.ResponseWriter, forceCheck bool) {
	if forceCheck {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
}

func acceptsGzip(r *http.Request) bool {
	for _, v := range r.Header.Values("Accept-Encoding") {
		if containsToken(v, "gzip") {
			return true
		}
	}
	return false
}

func containsToken(headerVal, token string) bool {
	// простой поиск подстроки с учётом запятых
	start := 0
	for i := 0; i <= len(headerVal); i++ {
		if i == len(headerVal) || headerVal[i] == ',' {
			part := trimSpaces(headerVal[start:i])
			// у "gzip;q=0" отбрасываем параметры
			for j := 0; j < len(part); j++ {
				if part[j] == ';' {
					part = part[:j]
					break
				}
			}
			part = trimSpaces(part)
			if part == token {
				return true
			}
			start = i + 1
		}
	}
	return false
}

func trimSpaces(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t') {
		j--
	}
	return s[i:j]
}
