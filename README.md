# Pharmacy System API

Это API для системы аптек, предоставляющее базовые CRUD-операции для работы с аптеками и лекарствами, с подключением к базе данных PostgreSQL. API позволяет управлять записями аптек и лекарств, а также выполняет миграцию для создания необходимых таблиц в базе данных.

## Структура проекта

Проект состоит из следующих частей:
- **API для аптек**: позволяет получить, добавить, обновить и удалить аптеки.
- **API для лекарств**: позволяет получить, добавить, обновить и удалить лекарства.
- **Миграции**: создают необходимые таблицы в базе данных при запуске приложения.

## Требования

- Go 1.18 или выше
- PostgreSQL (или Docker для запуска базы данных)
- Строки подключения к базе данных должны быть указаны в переменных окружения.

## Установка и запуск

### 1. Клонируйте репозиторий:

```bash
git clone https://github.com/your-username/pharmacy-system.git
cd pharmacy-system
```

### 2. Установите зависимости Go:

```bash
go mod tidy
```

### 3. Настройте переменные окружения

Перед запуском убедитесь, что переменные окружения для подключения к базе данных настроены:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=discorre
export DB_PASSWORD=0412
export DB_NAME=pharmacy_db
```

Если вы используете Docker для базы данных, вы можете создать контейнер PostgreSQL с помощью следующей команды:

```bash
docker run --name pharmacy_db -e POSTGRES_PASSWORD=0412 -e POSTGRES_USER=discorre -e POSTGRES_DB=pharmacy_db -p 5432:5432 -d postgres
```

### 4. Запустите приложение

После того как все переменные окружения настроены, запустите сервер с помощью команды:

```bash
go run main.go
```

API будет доступен по адресу [http://localhost:8080](http://localhost:8080).

### 5. Миграции

При запуске приложения будет выполнена миграция базы данных. Это создаст необходимые таблицы (если они еще не существуют) для работы с аптечными данными и лекарствами.

## API эндпоинты

### Аптеки:

- **GET** `/api/pharmacies` — Получить список всех аптек
- **GET** `/api/pharmacies/{id}` — Получить аптеку по ID
- **POST** `/api/pharmacies` — Создать новую аптеку
- **PUT** `/api/pharmacies/{id}` — Обновить информацию о аптеке
- **DELETE** `/api/pharmacies/{id}` — Удалить аптеку по ID

### Лекарства:

- **GET** `/api/medicines` — Получить список всех лекарств
- **GET** `/api/medicines/{id}` — Получить информацию о лекарстве по ID
- **POST** `/api/medicines` — Добавить новое лекарство
- **PUT** `/api/medicines/{id}` — Обновить информацию о лекарстве
- **DELETE** `/api/medicines/{id}` — Удалить лекарство по ID

## Тестирование API

Для тестирования API вы можете использовать инструменты, такие как **Postman** или **cURL**.

### Примеры запросов с cURL:

- Получить список всех аптек:

```bash
curl -X GET http://localhost:8080/api/pharmacies
```

- Получить аптеку по ID:

```bash
curl -X GET http://localhost:8080/api/pharmacies/1
```

- Создать новую аптеку:

```bash
curl -X POST http://localhost:8080/api/pharmacies -d '{"name": "Аптека №3", "address": "ул. Гагарина, 7"}' -H "Content-Type: application/json"
```

- Обновить аптеку:

```bash
curl -X PUT http://localhost:8080/api/pharmacies/1 -d '{"name": "Аптека №1 Обновленная", "address": "ул. Ленина, 10, Москва"}' -H "Content-Type: application/json"
```

- Удалить аптеку:

```bash
curl -X DELETE http://localhost:8080/api/pharmacies/1
```

## Структура данных

### Аптека (`Pharmacy`):
```json
{
  "id": 1,
  "name": "Аптека №1",
  "address": "ул. Ленина, 10, Москва"
}
```

### Лекарство (`Medicine`):
```json
{
  "id": 1,
  "name": "Парацетамол",
  "manufacturer": "Производитель A",
  "production_date": "2024-10-01",
  "packaging": "500 мг",
  "price": 150.00,
  "pharmacy_ids": [1, 2]
}
```

### Пояснения:
1. **Общие указания**:
   - Строки подключения к базе данных должны быть настроены через переменные окружения.
   - Если используется Docker для запуска PostgreSQL, предоставлена команда для его запуска.

2. **API эндпоинты**:
   - Подробное описание доступных маршрутов для работы с аптеками и лекарствами.

3. **Тестирование API**:
   - Примеры команд с использованием `curl` для взаимодействия с API, что полезно для тестирования.

4. **Структура данных**:
   - Пример структуры данных для аптеки и лекарства в JSON-формате, который используется в API.