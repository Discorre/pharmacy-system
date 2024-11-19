package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"pharmacy-system/internal/models"
	//"strconv"

	"github.com/gorilla/mux"
)

func ListDrugsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, description, price, stock FROM drugs")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var drugs []models.Drug
		for rows.Next() {
			var drug models.Drug
			if err := rows.Scan(&drug.ID, &drug.Name, &drug.Description, &drug.Price, &drug.Stock); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			drugs = append(drugs, drug)
		}

		json.NewEncoder(w).Encode(drugs)
	}
}

func AddDrugHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var drug models.Drug
		if err := json.NewDecoder(r.Body).Decode(&drug); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO drugs (name, description, price, stock) VALUES ($1, $2, $3, $4)",
			drug.Name, drug.Description, drug.Price, drug.Stock)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func GetDrugHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		var drug models.Drug
		err := db.QueryRow("SELECT id, name, description, price, stock FROM drugs WHERE id = $1", id).
			Scan(&drug.ID, &drug.Name, &drug.Description, &drug.Price, &drug.Stock)
		if err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		json.NewEncoder(w).Encode(drug)
	}
}

func DeleteDrugHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		_, err := db.Exec("DELETE FROM drugs WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
