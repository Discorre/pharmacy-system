package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"pharmacy-test/config"
	"pharmacy-test/models"
)

// CreatePharmacy создает новую аптеку
func CreatePharmacy(w http.ResponseWriter, r *http.Request) {
	var pharmacy models.Pharmacy
	if err := json.NewDecoder(r.Body).Decode(&pharmacy); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO pharmacies (name, address) VALUES ($1, $2) RETURNING id`
	err := config.DB.QueryRow(query, pharmacy.Name, pharmacy.Address).Scan(&pharmacy.ID)
	if err != nil {
		http.Error(w, "Failed to create pharmacy", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(pharmacy)
}

// GetAllPharmacies возвращает список всех аптек
func GetAllPharmacies(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(`SELECT id, name, address FROM pharmacies`)
	if err != nil {
		http.Error(w, "Failed to fetch pharmacies", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pharmacies []models.Pharmacy
	for rows.Next() {
		var p models.Pharmacy
		if err := rows.Scan(&p.ID, &p.Name, &p.Address); err != nil {
			http.Error(w, "Failed to parse pharmacies", http.StatusInternalServerError)
			return
		}
		pharmacies = append(pharmacies, p)
	}

	json.NewEncoder(w).Encode(pharmacies)
}

// CreateMedicine создает новое лекарство
func CreateMedicine(w http.ResponseWriter, r *http.Request) {
	var medicine models.Medicine
	if err := json.NewDecoder(r.Body).Decode(&medicine); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO medicines (name, manufacturer, production_date, packaging, price) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := config.DB.QueryRow(query, medicine.Name, medicine.Manufacturer, medicine.ProductionDate, medicine.Packaging, medicine.Price).Scan(&medicine.ID)
	if err != nil {
		http.Error(w, "Failed to create medicine", http.StatusInternalServerError)
		return
	}

	// Привязываем лекарства к аптекам
	for _, pharmacyID := range medicine.PharmacyIDs {
		_, err := config.DB.Exec(`INSERT INTO pharmacy_medicines (pharmacy_id, medicine_id) VALUES ($1, $2)`, pharmacyID, medicine.ID)
		if err != nil {
			http.Error(w, "Failed to link medicine to pharmacy", http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(medicine)
}

// GetMedicinesInPharmacy возвращает все лекарства, которые есть в аптеке по ее ID
func GetMedicinesInPharmacy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pharmacyID, _ := strconv.Atoi(vars["id"])

	rows, err := config.DB.Query(`
		SELECT m.id, m.name, m.manufacturer, m.production_date, m.packaging, m.price
		FROM medicines m
		JOIN pharmacy_medicines pm ON m.id = pm.medicine_id
		WHERE pm.pharmacy_id = $1
	`, pharmacyID)

	if err != nil {
		http.Error(w, "Failed to fetch medicines", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var medicines []models.Medicine
	for rows.Next() {
		var medicine models.Medicine
		if err := rows.Scan(&medicine.ID, &medicine.Name, &medicine.Manufacturer, &medicine.ProductionDate, &medicine.Packaging, &medicine.Price); err != nil {
			http.Error(w, "Failed to parse medicines", http.StatusInternalServerError)
			return
		}
		medicines = append(medicines, medicine)
	}

	json.NewEncoder(w).Encode(medicines)
}

// SearchMedicineInPharmacy проверяет, есть ли определенное лекарство в аптеке
func SearchMedicineInPharmacy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pharmacyID, _ := strconv.Atoi(vars["pharmacy_id"])
	medicineName := vars["medicine_name"]

	rows, err := config.DB.Query(`
		SELECT m.id, m.name, m.manufacturer, m.production_date, m.packaging, m.price
		FROM medicines m
		JOIN pharmacy_medicines pm ON m.id = pm.medicine_id
		WHERE pm.pharmacy_id = $1 AND m.name ILIKE $2
	`, pharmacyID, "%"+medicineName+"%")

	if err != nil {
		http.Error(w, "Failed to search medicine", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var medicines []models.Medicine
	for rows.Next() {
		var medicine models.Medicine
		if err := rows.Scan(&medicine.ID, &medicine.Name, &medicine.Manufacturer, &medicine.ProductionDate, &medicine.Packaging, &medicine.Price); err != nil {
			http.Error(w, "Failed to parse medicines", http.StatusInternalServerError)
			return
		}
		medicines = append(medicines, medicine)
	}

	json.NewEncoder(w).Encode(medicines)
}
