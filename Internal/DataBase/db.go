package DataBase

import (
	"database/sql"
	"log"
	"os"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	dir := "./DB"
    	if _, err := os.Stat(dir); os.IsNotExist(err) {
        	err := os.MkdirAll(dir, os.ModePerm)
        	if err != nil {
            		log.Fatal("Ошибка создания директории:", err)
        	}
    	}
	var err error
	DB, err = sql.Open("sqlite", "./DB/calculator.db")
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
