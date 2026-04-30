package repositories

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/database/providers"
)

type MessageRepository interface {
	InitSchema() error
	Create(msg *models.Message) error
	FindByStanzaID(stanzaID string) (*models.Message, error)
	StartPartitionManager()
	Close() error
}

type messageRepository struct {
	db *providers.Database
}

func NewMessageRepository(db *providers.Database) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Close() error {
	return r.db.Close()
}

func (r *messageRepository) InitSchema() error {
	var query string
	if r.db.Driver == "postgres" {
		query = `CREATE TABLE IF NOT EXISTS messages (
			stanzaId VARCHAR(255) NOT NULL,
			chatId VARCHAR(255) NOT NULL,
			senderJid VARCHAR(255) NOT NULL,
			messageText TEXT,
			quotedStanzaId VARCHAR(255),
			createdAt TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (stanzaId, createdAt)
		) PARTITION BY RANGE (createdAt);`
	} else {
		query = `CREATE TABLE IF NOT EXISTS messages (
			stanzaId VARCHAR(255) PRIMARY KEY,
			chatId VARCHAR(255) NOT NULL,
			senderJid VARCHAR(255) NOT NULL,
			messageText TEXT,
			quotedStanzaId VARCHAR(255),
			createdAt DATETIME DEFAULT CURRENT_TIMESTAMP
		);`
	}

	_, err := r.db.DB.Exec(query)
	return err
}

func (r *messageRepository) Create(msg *models.Message) error {
	var quotedID sql.NullString
	if msg.QuotedStanzaID != nil {
		quotedID.String = *msg.QuotedStanzaID
		quotedID.Valid = true
	}

	_, err := r.db.DB.Exec(r.db.GetQuery(
		"INSERT INTO messages (stanzaId, chatId, senderJid, messageText, quotedStanzaId, createdAt) VALUES (?, ?, ?, ?, ?, ?)",
	), msg.StanzaID, msg.ChatID, msg.SenderJID, msg.MessageText, quotedID, msg.CreatedAt)

	return err
}

func (r *messageRepository) FindByStanzaID(stanzaID string) (*models.Message, error) {
	row := r.db.DB.QueryRow(r.db.GetQuery(
		"SELECT stanzaId, chatId, senderJid, messageText, quotedStanzaId, createdAt FROM messages WHERE stanzaId = ?",
	), stanzaID)

	var msg models.Message
	var quotedID sql.NullString

	err := row.Scan(&msg.StanzaID, &msg.ChatID, &msg.SenderJID, &msg.MessageText, &quotedID, &msg.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if quotedID.Valid {
		msg.QuotedStanzaID = &quotedID.String
	}

	return &msg, nil
}

func (r *messageRepository) StartPartitionManager() {
	if r.db.Driver != "postgres" {
		slog.Info("Skipping partition manager (driver is not postgres)")
		return
	}

	r.managePartitions()

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			r.managePartitions()
		}
	}()
}

func (r *messageRepository) managePartitions() {
	slog.Info("Running daily partition manager for messages table")

	now := time.Now()
	daysToCreate := []time.Time{now, now.AddDate(0, 0, 1)}

	for _, day := range daysToCreate {
		nextDay := day.AddDate(0, 0, 1)
		tableName := fmt.Sprintf("messages_y%04dm%02dd%02d", day.Year(), day.Month(), day.Day())

		createSQL := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s PARTITION OF messages 
			FOR VALUES FROM ('%s') TO ('%s');`,
			tableName,
			day.Format("2006-01-02"),
			nextDay.Format("2006-01-02"))

		if _, err := r.db.DB.Exec(createSQL); err != nil {
			slog.Error("Failed to create partition", "table", tableName, "err", err)
		}
	}

	oldestDay := now.AddDate(0, 0, -15)
	oldestTable := fmt.Sprintf("messages_y%04dm%02dd%02d", oldestDay.Year(), oldestDay.Month(), oldestDay.Day())

	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s;", oldestTable)
	if _, err := r.db.DB.Exec(dropSQL); err != nil {
		slog.Error("Failed to drop old partition", "table", oldestTable, "err", err)
	} else {
		slog.Info("Cleaned up old partition if it existed", "table", oldestTable)
	}
}
