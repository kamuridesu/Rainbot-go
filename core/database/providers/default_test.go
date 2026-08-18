package providers

import "testing"

func TestGetQuerySqlite(t *testing.T) {
	db := &Database{Driver: "sqlite3"}
	query := "SELECT * FROM chat WHERE chatId = ? AND prefix = ?"

	if got := db.GetQuery(query); got != query {
		t.Errorf("GetQuery() = %q, want unchanged %q", got, query)
	}
}

func TestGetQueryPostgres(t *testing.T) {
	db := &Database{Driver: "postgres"}

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"no placeholders", "SELECT * FROM chat", "SELECT * FROM chat"},
		{"single placeholder", "SELECT * FROM chat WHERE chatId = ?", "SELECT * FROM chat WHERE chatId = $1"},
		{
			"multiple placeholders numbered in order",
			"UPDATE chat SET prefix = ?, adminOnly = ? WHERE chatId = ?",
			"UPDATE chat SET prefix = $1, adminOnly = $2 WHERE chatId = $3",
		},
		{"placeholder inside a longer clause", "DELETE FROM quotly WHERE chatId = ? AND fileId = ?", "DELETE FROM quotly WHERE chatId = $1 AND fileId = $2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := db.GetQuery(tt.query); got != tt.want {
				t.Errorf("GetQuery(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}

func TestGetQueryUnknownDriverReturnsUnchanged(t *testing.T) {
	db := &Database{Driver: "mysql"}
	query := "SELECT * FROM chat WHERE chatId = ?"

	if got := db.GetQuery(query); got != query {
		t.Errorf("GetQuery() = %q, want unchanged %q", got, query)
	}
}

func TestInitDBPanicsOnUnsupportedDriver(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected InitDB to panic for an unsupported driver")
		}
	}()
	InitDB("mysql", "irrelevant")
}

func TestInitDBSqlite(t *testing.T) {
	db, err := InitDB("sqlite3", "file:"+t.TempDir()+"/test.db")
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	defer db.Close()

	if db.Driver != "sqlite3" {
		t.Errorf("Driver = %q, want %q", db.Driver, "sqlite3")
	}
	if db.DB == nil {
		t.Error("expected a non-nil *sql.DB")
	}
}

func TestDatabaseCloseIsIdempotent(t *testing.T) {
	db, err := InitDB("sqlite3", "file:"+t.TempDir()+"/test.db")
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Errorf("second Close() error = %v, want nil (idempotent)", err)
	}
}
