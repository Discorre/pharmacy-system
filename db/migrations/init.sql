-- Таблица аптек
CREATE TABLE pharmacies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    address VARCHAR(255) UNIQUE NOT NULL,
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
