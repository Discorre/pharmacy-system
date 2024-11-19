-- Создание таблицы для хранения лекарств
CREATE TABLE IF NOT EXISTS drugs (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    stock INT NOT NULL
);

-- Создание таблицы для аптек
CREATE TABLE IF NOT EXISTS pharmacies (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL
);

-- Создание таблицы для связи лекарств с аптеками
CREATE TABLE IF NOT EXISTS pharmacy_stock (
    pharmacy_id INT NOT NULL,
    drug_id INT NOT NULL,
    stock INT NOT NULL,
    PRIMARY KEY (pharmacy_id, drug_id),
    FOREIGN KEY (pharmacy_id) REFERENCES pharmacies(id) ON DELETE CASCADE,
    FOREIGN KEY (drug_id) REFERENCES drugs(id) ON DELETE CASCADE
);
