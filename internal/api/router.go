package api

import (
	"database/sql"
	//"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(db *sql.DB) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/drugs", ListDrugsHandler(db)).Methods("GET")
	router.HandleFunc("/drugs", AddDrugHandler(db)).Methods("POST")
	router.HandleFunc("/drugs/{id}", GetDrugHandler(db)).Methods("GET")
	router.HandleFunc("/drugs/{id}", DeleteDrugHandler(db)).Methods("DELETE")

	return router
}
