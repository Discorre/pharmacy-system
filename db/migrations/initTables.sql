    -- Таблица адресов
    CREATE TABLE addresses (
        id SERIAL PRIMARY KEY,
        street VARCHAR(255) NOT NULL,
        city VARCHAR(255) NOT NULL,
        state VARCHAR(255),
        postal_code VARCHAR(20),
        country VARCHAR(100) NOT NULL
    );

    -- Таблица аптек
    CREATE TABLE pharmacies (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        address_id INT REFERENCES addresses(id) ON DELETE CASCADE
    );

    -- Таблица лекарств
    CREATE TABLE medicines (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255),
        manufacturer VARCHAR(255),
        production_date DATE,
        packaging VARCHAR(255),
        price NUMERIC(10, 2)
    );

    -- Связь аптек и лекарств (многие ко многим)
    CREATE TABLE pharmacy_medicines (
        pharmacy_id INT REFERENCES pharmacies(id) ON DELETE CASCADE,
        medicine_id INT REFERENCES medicines(id) ON DELETE CASCADE,
        PRIMARY KEY (pharmacy_id, medicine_id)
    );

    -- Таблица пользователей
    CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(255) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        cookie VARCHAR(255),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Дата создания
        last_login_at TIMESTAMP                         -- Дата последнего входа
    );

    -- Таблица деталей пользователей
    CREATE TABLE user_details (
        id SERIAL PRIMARY KEY,
        user_id INT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        first_name VARCHAR(255) NOT NULL,
        second_name VARCHAR(255) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        phone_number VARCHAR(20) UNIQUE NOT NULL,
        position VARCHAR(100)
    );
