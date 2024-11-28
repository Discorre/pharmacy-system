package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"database/sql"
	"pharmacy-test/config"
)

var DB *sql.DB

type Pharmacy struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Medicine struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	Manufacturer   string   `json:"manufacturer"`
	ProductionDate string   `json:"production_date"`
	Packaging      string   `json:"packaging"`
	Price          float64  `json:"price"`
	PharmacyIDs    []int    `json:"pharmacy_ids"`
}

// Пример хранилища данных
var pharmacies []Pharmacy
var medicines []Medicine

// Функция для подключения к базе данных
func connectToDB() (*sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to DB: %v", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("unable to ping DB: %v", err)
	}

	return db, nil
}

// Инициализация данных
func init() {
	pharmacies = []Pharmacy{
		{ID: 1, Name: "Аптека №1", Address: "ул. Ленина, 10, Москва"},
		{ID: 2, Name: "Аптека №2", Address: "ул. Пушкина, 5, Москва"},
	}

	medicines = []Medicine{
		{ID: 1, Name: "Парацетамол", Manufacturer: "Производитель A", ProductionDate: "2024-10-01", Packaging: "500 мг", Price: 150.00, PharmacyIDs: []int{1, 2}},
		{ID: 2, Name: "Ибупрофен", Manufacturer: "Производитель B", ProductionDate: "2024-09-15", Packaging: "200 мг", Price: 120.00, PharmacyIDs: []int{2}},
	}
}

// Получение списка аптек
func getPharmacies(w http.ResponseWriter, r *http.Request) {
	db, err := connectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name, address FROM pharmacies")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching pharmacies: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pharmacies []Pharmacy
	for rows.Next() {
		var pharmacy Pharmacy
		if err := rows.Scan(&pharmacy.ID, &pharmacy.Name, &pharmacy.Address); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}
		pharmacies = append(pharmacies, pharmacy)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pharmacies)
}

// Получение аптеки по ID
func getPharmacyByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	db, err := connectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var pharmacy Pharmacy
	err = db.QueryRow("SELECT id, name, address FROM pharmacies WHERE id = $1", id).Scan(&pharmacy.ID, &pharmacy.Name, &pharmacy.Address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching pharmacy: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pharmacy)
}

// Создание новой аптеки
func createPharmacy(w http.ResponseWriter, r *http.Request) {
	var pharmacy Pharmacy
	err := json.NewDecoder(r.Body).Decode(&pharmacy)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	db, err := connectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = db.QueryRow("INSERT INTO pharmacies(name, address) VALUES($1, $2) RETURNING id", pharmacy.Name, pharmacy.Address).Scan(&pharmacy.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting pharmacy: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pharmacy)
}

// Обновление информации о аптеке
func updatePharmacy(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedPharmacy Pharmacy
	err = json.NewDecoder(r.Body).Decode(&updatedPharmacy)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	db, err := connectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE pharmacies SET name = $1, address = $2 WHERE id = $3", updatedPharmacy.Name, updatedPharmacy.Address, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating pharmacy: %v", err), http.StatusInternalServerError)
		return
	}

	updatedPharmacy.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPharmacy)
}

// Удаление аптеки
func deletePharmacy(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	db, err := connectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM pharmacies WHERE id = $1", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting pharmacy: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Получение списка лекарств
func getMedicines(w http.ResponseWriter, r *http.Request) {
	db, err := connectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name, manufacturer, production_date, packaging, price FROM medicines")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching medicines: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var medicines []Medicine
	for rows.Next() {
		var medicine Medicine
		if err := rows.Scan(&medicine.ID, &medicine.Name, &medicine.Manufacturer, &medicine.ProductionDate, &medicine.Packaging, &medicine.Price); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}
		medicines = append(medicines, medicine)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(medicines)
}

// Получение лекарства по ID
func getMedicineByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	db, err := connectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var medicine Medicine
	err = db.QueryRow("SELECT id, name, manufacturer, production_date, packaging, price FROM medicines WHERE id = $1", id).Scan(&medicine.ID, &medicine.Name, &medicine.Manufacturer, &medicine.ProductionDate, &medicine.Packaging, &medicine.Price)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching medicine: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(medicine)
}

// Создание нового лекарства
func createMedicine(w http.ResponseWriter, r *http.Request) {
	var medicine Medicine
	err := json.NewDecoder(r.Body).Decode(&medicine)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	db, err := connectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = db.QueryRow("INSERT INTO medicines(name, manufacturer, production_date, packaging, price) VALUES($1, $2, $3, $4, $5) RETURNING id", medicine.Name, medicine.Manufacturer, medicine.ProductionDate, medicine.Packaging, medicine.Price).Scan(&medicine.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting medicine: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(medicine)
}

func main() {
	// Устанавливаем подключение к базе данных
	var err error
	DB, err = sql.Open("postgres", "postgres://username:password@localhost:5432/pharmacy_db?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer DB.Close()

	// Выполняем миграцию
	config.InitDB()

	r := mux.NewRouter()

	r.HandleFunc("/api/pharmacies", getPharmacies).Methods("GET")
	r.HandleFunc("/api/pharmacies/{id:[0-9]+}", getPharmacyByID).Methods("GET")
	r.HandleFunc("/api/pharmacies", createPharmacy).Methods("POST")
	r.HandleFunc("/api/pharmacies/{id:[0-9]+}", updatePharmacy).Methods("PUT")
	r.HandleFunc("/api/pharmacies/{id:[0-9]+}", deletePharmacy).Methods("DELETE")

	r.HandleFunc("/api/medicines", getMedicines).Methods("GET")
	r.HandleFunc("/api/medicines/{id:[0-9]+}", getMedicineByID).Methods("GET")
	r.HandleFunc("/api/medicines", createMedicine).Methods("POST")

	log.Println("API сервер запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
