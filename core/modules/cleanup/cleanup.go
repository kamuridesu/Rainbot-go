package cleanup

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

	"github.com/kamuridesu/rainbot-go/core/services"
	"github.com/kamuridesu/rainbot-go/internal/storage"
)

// DefaultRetention is how long a chat's data is kept after the bot leaves it.
const DefaultRetention = 7 * 24 * time.Hour

type Cleaner struct {
	Chat      *services.ChatService
	Member    *services.MemberService
	Filter    *services.FilterService
	Quotly    *services.QuotlyService
	Storage   storage.FileStorage
	Retention time.Duration
	Now       func() time.Time
}

func New(chat *services.ChatService, member *services.MemberService, filter *services.FilterService, quotly *services.QuotlyService, fileStorage storage.FileStorage) *Cleaner {
	return &Cleaner{
		Chat:      chat,
		Member:    member,
		Filter:    filter,
		Quotly:    quotly,
		Storage:   fileStorage,
		Retention: DefaultRetention,
		Now:       time.Now,
	}
}

// Start runs the cleanup once right away and then every interval until ctx is done.
func (c *Cleaner) Start(ctx context.Context, interval time.Duration) {
	go func() {
		c.RunOnce(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.RunOnce(ctx)
			}
		}
	}()
}

// RunOnce purges every chat the bot left more than Retention ago and returns how many were purged.
func (c *Cleaner) RunOnce(ctx context.Context) int {
	cutoff := c.Now().Add(-c.Retention)

	if removed, err := c.Member.DeleteLeftBefore(cutoff); err != nil {
		slog.Error("Failed to clean up members who left", "err", err)
	} else if removed > 0 {
		slog.Info("Cleaned up members who left their chats", "count", removed)
	}

	chatIds, err := c.Chat.GetLeftBefore(cutoff)
	if err != nil {
		slog.Error("Failed to list chats pending cleanup", "err", err)
		return 0
	}

	purged := 0
	for _, chatId := range chatIds {
		if ctx.Err() != nil {
			break
		}
		ok, err := c.purgeChat(ctx, chatId, cutoff)
		if err != nil {
			slog.Error("Failed to clean up chat", "chat", chatId, "err", err)
			continue
		}
		if ok {
			slog.Info("Cleaned up data from chat the bot left", "chat", chatId)
			purged++
		}
	}
	return purged
}

func (c *Cleaner) purgeChat(ctx context.Context, chatId string, cutoff time.Time) (bool, error) {
	// Re-check right before deleting media in case the bot rejoined since the listing.
	chat, err := c.Chat.Get(chatId)
	if err != nil {
		return false, err
	}
	if chat == nil || chat.LeftAt == nil || chat.LeftAt.After(cutoff) {
		return false, nil
	}

	files, err := c.mediaFiles(chatId)
	if err != nil {
		return false, err
	}

	// Media goes first: if a delete fails, the database rows that reference it
	// are kept so the next run can retry instead of leaving orphaned files.
	for _, file := range files {
		if err := c.Storage.Delete(ctx, file); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return false, fmt.Errorf("deleting file %s: %w", file, err)
		}
	}

	return c.Chat.Purge(chatId, cutoff)
}

func (c *Cleaner) mediaFiles(chatId string) ([]string, error) {
	var files []string

	filters, err := c.Filter.GetFilters(chatId)
	if err != nil {
		return nil, err
	}
	for _, filter := range filters {
		if filter.Kind != "text" && filter.Response != "" {
			files = append(files, filter.Response)
		}
	}

	quotlies, err := c.Quotly.GetAllByChat(chatId)
	if err != nil {
		return nil, err
	}
	for _, quotly := range quotlies {
		files = append(files, quotly.FileId)
	}

	return files, nil
}
