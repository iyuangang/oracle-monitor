package storage

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
    db   *sql.DB
    mu   sync.RWMutex
}
type MonitorResult struct {
    DBName       string
    Status       bool
    ResponseTime float64
    Error        string
    CheckTime    time.Time
}

func NewStorage() *Storage {
    db, err := sql.Open("sqlite3", "monitor.db")
    if err != nil {
        panic(err)
    }

    storage := &Storage{db: db}
    storage.initDB()
    return storage
}

func (s *Storage) initDB() {
    s.mu.Lock()
    defer s.mu.Unlock()

    _, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS monitor_history (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            db_name TEXT NOT NULL,
            status INTEGER NOT NULL,
            response_time REAL,
            error TEXT,
            check_time TIMESTAMP NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_monitor_history_db_name 
        ON monitor_history(db_name);
        CREATE INDEX IF NOT EXISTS idx_monitor_history_check_time 
        ON monitor_history(check_time);
    `)

    if err != nil {
        panic(err)
    }
}

func (s *Storage) SaveResult(result MonitorResult) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    _, err := s.db.Exec(`
        INSERT INTO monitor_history 
        (db_name, status, response_time, error, check_time)
        VALUES (?, ?, ?, ?, ?)
    `,
        result.DBName,
        result.Status,
        result.ResponseTime,
        result.Error,
        result.CheckTime,
    )

    return err
}

func (s *Storage) CleanupOldData(days int) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    cutoff := time.Now().AddDate(0, 0, -days)
    _, err := s.db.Exec("DELETE FROM monitor_history WHERE check_time < ?", cutoff)
    return err
}
