package DataBase

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./calculator.db")
	if err != nil {
		log.Fatal("Ошибка подключения к SQLite:", err)
	}
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            login TEXT UNIQUE,
            password_hash TEXT
        )
    `)
	if err != nil {
		log.Fatal("Ошибка создания таблицы users:", err)
	}
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS expressions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER,
            expression TEXT,
            status TEXT,
            result REAL,
            FOREIGN KEY (user_id) REFERENCES users(id)
        )
    `)
	if err != nil {
		log.Fatal("Ошибка создания таблицы expressions:", err)
	}
}
