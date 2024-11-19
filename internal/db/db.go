package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

const (
	connStr = "postgres://discorre:0412@localhost:5432/pharmacy?sslmode=disable"
)

func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Успешное подключение к базе данных")
	return db, nil
}
