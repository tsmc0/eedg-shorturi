package models

import "time"

// CreateRequest — входной JSON для создания шортлинка.
type CreateRequest struct {
	URL          string `json:"url"`
	CacheContent bool   `json:"cache_content"`
	TTLSeconds   int    `json:"ttl_seconds,omitempty"`
	ForceCheck   bool   `json:"force_check,omitempty"`
}

// CreateResponse — ответ на создание.
type CreateResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

// LinkRow — запись в БД.
type LinkRow struct {
	Code       string
	LongURL    string
	ForceCheck bool
	Cache      []byte // gzip-сжатый контент
	MimeType   string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

// InfoResponse — метаданные по ссылке.
type InfoResponse struct {
	Code       string    `json:"code"`
	LongURL    string    `json:"long_url"`
	ForceCheck bool      `json:"force_check"`
	HasCache   bool      `json:"has_cache"`
	CacheSize  int       `json:"cache_size"`
	Mime       string    `json:"mime,omitempty"`
	ExpiresAt  time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// RefreshResponse — ответ обновления кеша.
type RefreshResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Mime      string    `json:"mime"`
	SizeGz    int       `json:"size_gz"`
}

// ErrorResponse — универсальный ответ ошибки.
type ErrorResponse struct {
	Error string `json:"error"`
}
