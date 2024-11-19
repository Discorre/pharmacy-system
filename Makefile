# Переменные
APP_NAME = pharmacy-system
DB_USER = discorre
DB_PASSWORD = 0412
DB_NAME = pharmacy
DB_HOST = localhost
DB_PORT = 5432

# Команды
.PHONY: all build run migrate clean db-start db-stop db-reset

# Сборка проекта
build:
	@echo "Компиляция проекта..."
	go build -o $(APP_NAME) ./cmd/main.go

# Запуск проекта
run: build
	@echo "Запуск приложения..."
	./$(APP_NAME)

# Установка зависимостей
deps:
	@echo "Установка зависимостей..."
	go mod tidy

# Выполнение миграций
migrate:
	@echo "Выполнение миграций..."
	PGPASSWORD=$(DB_PASSWORD) psql -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) -d $(DB_NAME) -f internal/db/migration.sql

# Подготовка базы данных (создание)
db-start:
	@echo "Создание базы данных..."
	PGPASSWORD=$(DB_PASSWORD) createdb -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) $(DB_NAME)

# Удаление базы данных
db-stop:
	@echo "Удаление базы данных..."
	PGPASSWORD=$(DB_PASSWORD) dropdb -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) --if-exists $(DB_NAME)

# Полный сброс базы данных
db-reset: db-stop db-start migrate

# Очистка собранного бинарного файла
clean:
	@echo "Очистка проекта..."
	rm -f $(APP_NAME)