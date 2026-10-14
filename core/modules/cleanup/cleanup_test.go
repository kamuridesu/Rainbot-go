package cleanup

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/providers"
	"github.com/kamuridesu/rainbot-go/core/database/repositories"
	"github.com/kamuridesu/rainbot-go/core/services"
	"github.com/kamuridesu/rainbot-go/internal/utils"
)

type fakeStorage struct {
	deleted   []string
	deleteErr map[string]error
}

func (f *fakeStorage) Exists(ctx context.Context, filename string) (bool, error) { return true, nil }
func (f *fakeStorage) Read(ctx context.Context, filename string) ([]byte, error) { return nil, nil }
func (f *fakeStorage) Write(ctx context.Context, filename string, data []byte) error {
	return nil
}
func (f *fakeStorage) Delete(ctx context.Context, filename string) error {
	if err := f.deleteErr[filename]; err != nil {
		return err
	}
	f.deleted = append(f.deleted, filename)
	return nil
}

var now = time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

func newTestCleaner(t *testing.T) (*Cleaner, *fakeStorage, repositories.ChatRepository, *providers.Database) {
	t.Helper()

	_, file, _, _ := runtime.Caller(0)
	dbfile := filepath.Join(t.TempDir(), "test.db")
	db, err := providers.InitDB("sqlite3", "file:"+dbfile+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	t.Chdir(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	if err := utils.MigrateSqlite(db); err != nil {
		t.Fatalf("MigrateSqlite() error = %v", err)
	}

	chatRepo := repositories.NewChatRepository(db)
	memberRepo := repositories.NewMemberRepository(db)
	filterRepo := repositories.NewFilterRepository(db)
	quotlyRepo := repositories.NewQuotlyRepository(db)

	store := &fakeStorage{deleteErr: map[string]error{}}
	c := New(
		services.NewChatService(chatRepo),
		services.NewMemberService(memberRepo),
		services.NewFilterRepository(filterRepo),
		services.NewQuotlyService(quotlyRepo),
		store,
	)
	c.Now = func() time.Time { return now }
	return c, store, chatRepo, db
}

func seedChat(t *testing.T, db *providers.Database, chatRepo repositories.ChatRepository, id string, leftAgo time.Duration) {
	t.Helper()
	if err := chatRepo.Create(id); err != nil {
		t.Fatalf("Create(%s) error = %v", id, err)
	}
	filterRepo := repositories.NewFilterRepository(db)
	quotlyRepo := repositories.NewQuotlyRepository(db)
	filterRepo.Create(&models.Filter{ChatID: id, Pattern: "img", Kind: "image", Response: id + ".jpg"})
	filterRepo.Create(&models.Filter{ChatID: id, Pattern: "txt", Kind: "text", Response: "hello"})
	quotlyRepo.Create(&models.QuotlyFile{ChatID: id, FileId: id + ".png"})
	if leftAgo > 0 {
		chatRepo.MarkLeft(id, now.Add(-leftAgo))
	}
}

func TestRunOncePurgesChatsLeftMoreThanAWeekAgo(t *testing.T) {
	c, store, chatRepo, db := newTestCleaner(t)
	seedChat(t, db, chatRepo, "old@g.us", 8*24*time.Hour)
	seedChat(t, db, chatRepo, "recent@g.us", 6*24*time.Hour)
	seedChat(t, db, chatRepo, "active@g.us", 0)

	if purged := c.RunOnce(context.Background()); purged != 1 {
		t.Errorf("RunOnce() = %d, want 1", purged)
	}

	slices.Sort(store.deleted)
	want := []string{"old@g.us.jpg", "old@g.us.png"}
	if !slices.Equal(store.deleted, want) {
		t.Errorf("deleted files = %v, want %v", store.deleted, want)
	}

	if chat, _ := chatRepo.FindById("old@g.us"); chat != nil {
		t.Error("old chat should have been purged")
	}
	for _, id := range []string{"recent@g.us", "active@g.us"} {
		if chat, _ := chatRepo.FindById(id); chat == nil {
			t.Errorf("%s should not have been purged", id)
		}
	}
}

func TestRunOnceIgnoresAlreadyMissingFiles(t *testing.T) {
	c, store, chatRepo, db := newTestCleaner(t)
	seedChat(t, db, chatRepo, "old@g.us", 8*24*time.Hour)
	store.deleteErr["old@g.us.jpg"] = &fs.PathError{Op: "remove", Path: "old@g.us.jpg", Err: fs.ErrNotExist}

	if purged := c.RunOnce(context.Background()); purged != 1 {
		t.Errorf("RunOnce() = %d, want 1", purged)
	}
}

func TestRunOnceKeepsRowsWhenFileDeleteFails(t *testing.T) {
	c, store, chatRepo, db := newTestCleaner(t)
	seedChat(t, db, chatRepo, "old@g.us", 8*24*time.Hour)
	store.deleteErr["old@g.us.png"] = errors.New("s3 unavailable")

	if purged := c.RunOnce(context.Background()); purged != 0 {
		t.Errorf("RunOnce() = %d, want 0", purged)
	}
	if chat, _ := chatRepo.FindById("old@g.us"); chat == nil {
		t.Fatal("chat rows should be kept so the next run can retry")
	}

	delete(store.deleteErr, "old@g.us.png")
	if purged := c.RunOnce(context.Background()); purged != 1 {
		t.Errorf("retry RunOnce() = %d, want 1", purged)
	}
}

func TestRunOnceRemovesMembersWhoLeftButKeepsQuotes(t *testing.T) {
	c, store, chatRepo, db := newTestCleaner(t)
	seedChat(t, db, chatRepo, "active@g.us", 0)
	memberRepo := repositories.NewMemberRepository(db)
	memberRepo.Create("active@g.us", "old@lid")
	memberRepo.Create("active@g.us", "recent@lid")
	memberRepo.MarkLeft("active@g.us", "old@lid", now.Add(-8*24*time.Hour))
	memberRepo.MarkLeft("active@g.us", "recent@lid", now.Add(-24*time.Hour))

	c.RunOnce(context.Background())

	if m, _ := memberRepo.FindByChatAndId("active@g.us", "old@lid"); m != nil {
		t.Error("member who left over a week ago should be deleted")
	}
	if m, _ := memberRepo.FindByChatAndId("active@g.us", "recent@lid"); m == nil {
		t.Error("member who left recently should be kept")
	}
	if len(store.deleted) != 0 {
		t.Errorf("deleted files = %v, want none", store.deleted)
	}
	quotes, _ := repositories.NewQuotlyRepository(db).FindAllByChat("active@g.us")
	if len(quotes) != 1 {
		t.Errorf("quotes = %d, want 1", len(quotes))
	}
}
