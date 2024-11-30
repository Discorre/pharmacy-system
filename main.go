package main

import (
	"log"
	"net/http"

	"database/sql"
	"pharmacy-test/config"
	"pharmacy-test/handlers"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var DB *sql.DB

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

	r.HandleFunc("/api/pharmacies", handlers.GetPharmacies).Methods("GET")
	r.HandleFunc("/api/pharmacies/{id:[0-9]+}", handlers.GetPharmacyByID).Methods("GET")
	r.HandleFunc("/api/pharmacies", handlers.CreatePharmacy).Methods("POST")
	r.HandleFunc("/api/pharmacies/{id:[0-9]+}", handlers.UpdatePharmacy).Methods("PUT")
	r.HandleFunc("/api/pharmacies/{id:[0-9]+}", handlers.DeletePharmacy).Methods("DELETE")

	r.HandleFunc("/api/medicines", handlers.GetMedicines).Methods("GET")
	r.HandleFunc("/api/medicines/{Aid:[0-9]+}", handlers.GetMedicineByID).Methods("GET")
	r.HandleFunc("/api/medicines", handlers.CreateMedicine).Methods("POST")
	r.HandleFunc("/api/medicines/{id:[0-9]+}", handlers.UpdateMedicine).Methods("PUT")
	r.HandleFunc("/api/medicines/{id:[0-9]+}", handlers.DeleteMedicine).Methods("DELETE")

	// Маршруты для пользователей
	r.HandleFunc("/api/users", handlers.CreateUserWithDetails).Methods("POST")
	r.HandleFunc("/api/users/{id}", handlers.UpdateUserWithDetails).Methods("PUT")
	r.HandleFunc("/api/users/{id}", handlers.GetUserWithDetailsByID).Methods("GET")
	r.HandleFunc("/api/users/{id}", handlers.DeleteUserWithDetails).Methods("DELETE")
	r.HandleFunc("/api/users", handlers.GetAllUsersWithDetails).Methods("GET")

	// Маршруты для аутентификации и авторизации
	r.HandleFunc("/api/users/login", handlers.LoginUser).Methods("POST")

	// Доступ к таблицам аптек и лекарств
	r.HandleFunc("/api/pharmacies", handlers.RoleMiddleware("Seller", handlers.GetPharmacies)).Methods("GET")
	r.HandleFunc("/api/pharmacies/{id:[0-9]+}", handlers.RoleMiddleware("Seller", handlers.GetPharmacyByID)).Methods("GET")

	// Только для создания, обновления и удаления
	r.HandleFunc("/api/pharmacies", handlers.RoleMiddleware("Seller", handlers.CreatePharmacy)).Methods("POST")
	r.HandleFunc("/api/pharmacies/{id:[0-9]+}", handlers.RoleMiddleware("Seller", handlers.UpdatePharmacy)).Methods("PUT")
	r.HandleFunc("/api/pharmacies/{id:[0-9]+}", handlers.RoleMiddleware("Seller", handlers.DeletePharmacy)).Methods("DELETE")

	// Управление лекарствами доступно только для Seller и Developer
	r.HandleFunc("/api/medicines", handlers.RoleMiddleware("Seller", handlers.GetMedicines)).Methods("GET")
	r.HandleFunc("/api/medicines/{Aid:[0-9]+}", handlers.RoleMiddleware("Seller", handlers.GetMedicineByID)).Methods("GET")
	r.HandleFunc("/api/medicines", handlers.RoleMiddleware("Seller", handlers.CreateMedicine)).Methods("POST")
	r.HandleFunc("/api/medicines/{id:[0-9]+}", handlers.RoleMiddleware("Seller", handlers.UpdateMedicine)).Methods("PUT")
	r.HandleFunc("/api/medicines/{id:[0-9]+}", handlers.RoleMiddleware("Seller", handlers.DeleteMedicine)).Methods("DELETE")




	log.Println("API сервер запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", handlers.EnableCORS(r)))
}
