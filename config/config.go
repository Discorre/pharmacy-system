package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	// Строка подключения к PostgreSQL с использованием административных прав (например, база данных "postgres")
	adminConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "discorre"), // Здесь используем администратора "discorre"
		getEnv("DB_PASSWORD", "0412"),
	)

	// Подключаемся с правами администратора к базе данных "postgres"
	adminDB, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		log.Fatalf("Unable to connect to DB as admin: %v", err)
	}
	defer adminDB.Close()

	// Проверка существования базы данных
	var exists bool
	err = adminDB.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", getEnv("DB_NAME", "pharmacy_system"))).Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking database existence: %v", err)
	}

	// Если база данных не существует, создаем её
	if !exists {
		_, err := adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", getEnv("DB_NAME", "pharmacy_system")))
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		log.Printf("Database %s created successfully!", getEnv("DB_NAME", "pharmacy_system"))
	}

	// Теперь подключаемся к базе данных "pharmacy_system"
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "discorre"),
		getEnv("DB_PASSWORD", "0412"),
		getEnv("DB_NAME", "pharmacy_system"),
	)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v", err)
	}

	// Проверка соединения с базой данных
	if err = DB.Ping(); err != nil {
		log.Fatalf("Unable to ping DB: %v", err)
	}

	log.Println("Connected to the database!")

	// Применяем миграции для создания таблиц
	createTablesIfNotExist()
}

func createTablesIfNotExist() {
	// Проверка и создание таблицы pharmacies
	var exists bool
	err := DB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'pharmacies'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking pharmacies table existence: %v", err)
	}

	if !exists {
		createPharmaciesTable := `
		CREATE TABLE pharmacies (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255),
			address VARCHAR(255) UNIQUE NOT NULL
		);
		`
		_, err := DB.Exec(createPharmaciesTable)
		if err != nil {
			log.Fatalf("Failed to create pharmacies table: %v", err)
		}
		log.Println("pharmacies table created successfully!")
	}

	// Проверка и создание таблицы medicines
	err = DB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'medicines'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking medicines table existence: %v", err)
	}

	if !exists {
		createMedicinesTable := `
		CREATE TABLE medicines (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255),
			manufacturer VARCHAR(255),
			production_date DATE,
			packaging VARCHAR(255),
			price NUMERIC(10, 2)
		);
		`
		_, err := DB.Exec(createMedicinesTable)
		if err != nil {
			log.Fatalf("Failed to create medicines table: %v", err)
		}
		log.Println("medicines table created successfully!")
	}

	// Проверка и создание таблицы pharmacy_medicines
	err = DB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'pharmacy_medicines'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking pharmacy_medicines table existence: %v", err)
	}

	if !exists {
		createManyToManyTable := `
		CREATE TABLE pharmacy_medicines (
			pharmacy_id INT REFERENCES pharmacies(id) ON DELETE CASCADE,
			medicine_id INT REFERENCES medicines(id) ON DELETE CASCADE,
			PRIMARY KEY (pharmacy_id, medicine_id)
		);
		`
		_, err := DB.Exec(createManyToManyTable)
		if err != nil {
			log.Fatalf("Failed to create pharmacy_medicines table: %v", err)
		}
		log.Println("pharmacy_medicines table created successfully!")
	}

	log.Println("All tables checked and created if necessary!")
}

// Функция для получения переменной окружения с значением по умолчанию
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
