package handlers

import (
	"time"

	"eedg-shortlink/storage"
)

// App — общие зависимости хендлеров.
type App struct {
	Store        *storage.Store
	DefaultTTL   time.Duration
	BaseShortURL string
	FetchTimeout time.Duration
	MaxContentMB int64
}

// maxContentBytes возвращает лимит скачиваемого контента в байтах.
func (a *App) maxContentBytes() int64 {
	if a.MaxContentMB <= 0 {
		return 5 << 20
	}
	return a.MaxContentMB << 20
}
