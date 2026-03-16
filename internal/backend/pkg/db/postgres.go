package db
import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)
func NewDatabase(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("veritabanı bağlantısı açılamadı: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("veritabanına erişilemiyor: %w", err)
	}
	return db, nil
}
