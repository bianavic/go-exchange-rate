package config

import "database/sql"

const (
	dbPath      = "cotacoes.db"
	createQuery = `
	CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
)

func InitDB() (*sql.DB, error) {
	logger := GetLogger("sqlite")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Errorf("error opening sqlite: %v", err)
		return nil, err
	}

	_, err = db.Exec(createQuery)
	if err != nil {
		logger.Errorf("failed to create table: %v", err)
		return nil, err
	}

	return db, nil
}
