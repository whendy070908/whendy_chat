package internal

import "database/sql"
import _ "modernc.org/sqlite"

const DBPath = "./whendy_chat.db"

func OpenDB() *sql.DB {
	db, _ := sql.Open("sqlite", DBPath)
	return db
}

func Migrate(db *sql.DB) {
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY,
			username TEXT UNIQUE,
			password TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS servers(
			id INTEGER PRIMARY KEY,
			name TEXT,
			owner_id INTEGER,
			created_at INTEGER
		);`,
		`CREATE TABLE IF NOT EXISTS server_members(
			server_id INTEGER,
			user_id INTEGER,
			role TEXT,
			PRIMARY KEY(server_id,user_id)
		);`,
		`CREATE TABLE IF NOT EXISTS channels(
			id INTEGER PRIMARY KEY,
			server_id INTEGER,
			name TEXT,
			created_at INTEGER
		);`,
		`CREATE TABLE IF NOT EXISTS messages(
			id INTEGER PRIMARY KEY,
			channel_id INTEGER,
			user_id INTEGER,
			content TEXT,
			created_at INTEGER
		);`,
	}
	for _, s := range sqls {
		db.Exec(s)
	}
}
