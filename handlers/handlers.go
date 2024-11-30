package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
)

var DB *sql.DB

// Address represents an address.
type Address struct {
	ID         int    `json:"id"`
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country"`
}

// Pharmacy represents a pharmacy with an address.
type Pharmacy struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Address Address `json:"address"`
}

// Medicine represents a medicine with associated pharmacies.
type Medicine struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	Manufacturer   string   `json:"manufacturer"`
	ProductionDate string   `json:"production_date"`
	Packaging      string   `json:"packaging"`
	Price          float64  `json:"price"`
	PharmacyIDs    []int    `json:"pharmacy_ids"`
}

type UserWithDetails struct {
	ID        int           `json:"id"`
	Username  string        `json:"username"`
	Password  string        `json:"password"`
	Cookie	  string    	`json:"cookie"`
	CreatedAt time.Time     `json:"created_at"`
	Details   UserDetails   `json:"details"`
}

type UserDetails struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	FirstName   string `json:"first_name"`
	SecondName  string `json:"second_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Position    string `json:"position"`
}

// LoginRequest структура для получения данных из тела запроса
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse структура для отправки данных клиенту
type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}



func EnableCORS(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if r.Method == "OPTIONS" {
            return
        }
        h.ServeHTTP(w, r)
    })
}


// Функция для подключения к базе данных
func ConnectToDB() (*sql.DB, error) {
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

// Хэширование пароля
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Функция для проверки валидности позиции
func isValidPosition(position string) bool {
	validPositions := []string{"Developer", "Seller", "Buyer"}
	for _, pos := range validPositions {
		if position == pos {
			return true
		}
	}
	return false
}

// LoginUser обрабатывает вход пользователя
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Чтение JSON из запроса
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var user UserWithDetails

	// Получение данных пользователя из базы
	err = db.QueryRow("SELECT u.id, u.username, u.password, u.cookie, ud.position FROM users u JOIN user_details ud ON u.id = ud.user_id WHERE u.username = $1",
		credentials.Username).Scan(&user.ID, &user.Username, &user.Password, &user.Cookie, &user.Details.Position)
	if err != nil {
		http.Error(w, `{"message": "Invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		http.Error(w, `{"message": "Invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	// Сохранение cookie на клиенте
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    user.Cookie,
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	// Возврат роли и данных пользователя
	response := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"position": user.Details.Position,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func RoleMiddleware(requiredPosition string, handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("auth_token")
        if err != nil {
            http.Error(w, `{"message": "Unauthorized"}`, http.StatusUnauthorized)
            return
        }

        db, err := ConnectToDB()
        if err != nil {
            http.Error(w, `{"message": "Database error"}`, http.StatusInternalServerError)
            return
        }
        defer db.Close()

        var userPosition string
        err = db.QueryRow("SELECT ud.position FROM users u JOIN user_details ud ON u.id = ud.user_id WHERE u.cookie = $1", cookie.Value).Scan(&userPosition)
        if err != nil {
            http.Error(w, `{"message": "Invalid user session"}`, http.StatusUnauthorized)
            return
        }

        if userPosition != requiredPosition && (requiredPosition != "Seller" || userPosition != "Buyer") {
            http.Error(w, `{"message": "Forbidden"}`, http.StatusForbidden)
            return
        }

        handler(w, r)
    }
}


func CreateUserWithDetails(w http.ResponseWriter, r *http.Request) {
	var userWithDetails UserWithDetails
	err := json.NewDecoder(r.Body).Decode(&userWithDetails)
	if err != nil {
		http.Error(w, `{"message": "Invalid input"}`, http.StatusBadRequest)
		return
	}

	// Проверка на валидность позиции
	if !isValidPosition(userWithDetails.Details.Position) {
		http.Error(w, `{"message": "Invalid position"}`, http.StatusBadRequest)
		return
	}

	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "Error connecting to DB: %v"}`, err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Хэширование пароля
	hashedPassword, err := hashPassword(userWithDetails.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "Error hashing password: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Создание UUID
	cookie := uuid.New().String()

	// Начало транзакции
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "Error starting transaction: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Вставка пользователя
	err = tx.QueryRow("INSERT INTO users(username, password, cookie) VALUES($1, $2, $3) RETURNING id, created_at", userWithDetails.Username, hashedPassword, cookie).Scan(&userWithDetails.ID, &userWithDetails.CreatedAt)
	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf(`{"message": "Error inserting user: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Вставка деталей пользователя
	userWithDetails.Details.UserID = userWithDetails.ID
	err = tx.QueryRow("INSERT INTO user_details(user_id, first_name, second_name, email, phone_number, position) VALUES($1, $2, $3, $4, $5, $6) RETURNING id",
		userWithDetails.Details.UserID, userWithDetails.Details.FirstName, userWithDetails.Details.SecondName, userWithDetails.Details.Email, userWithDetails.Details.PhoneNumber, userWithDetails.Details.Position).Scan(&userWithDetails.Details.ID)
	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf(`{"message": "Error inserting user details: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Подтверждение транзакции
	err = tx.Commit()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "Error committing transaction: %v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userWithDetails)
}

func UpdateUserWithDetails(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var userWithDetails UserWithDetails
	err = json.NewDecoder(r.Body).Decode(&userWithDetails)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Проверка на валидность позиции
	if !isValidPosition(userWithDetails.Details.Position) {
		http.Error(w, "Invalid position", http.StatusBadRequest)
		return
	}

	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Хэширование пароля
	hashedPassword, err := hashPassword(userWithDetails.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error hashing password: %v", err), http.StatusInternalServerError)
		return
	}

	// Начало транзакции
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Обновление пользователя
	_, err = tx.Exec("UPDATE users SET username = $1, password = $2 WHERE id = $3", userWithDetails.Username, hashedPassword, id)
	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error updating user: %v", err), http.StatusInternalServerError)
		return
	}

	// Обновление деталей пользователя
	_, err = tx.Exec("UPDATE user_details SET first_name = $1, second_name = $2, email = $3, phone_number = $4, position = $5 WHERE user_id = $6",
		userWithDetails.Details.FirstName, userWithDetails.Details.SecondName, userWithDetails.Details.Email, userWithDetails.Details.PhoneNumber, userWithDetails.Details.Position, id)
	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error updating user details: %v", err), http.StatusInternalServerError)
		return
	}

	// Подтверждение транзакции
	err = tx.Commit()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error committing transaction: %v", err), http.StatusInternalServerError)
		return
	}

	userWithDetails.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userWithDetails)
}

// Получение пользователя и его деталей по ID
func GetUserWithDetailsByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var userWithDetails UserWithDetails
	err = db.QueryRow("SELECT id, username, created_at FROM users WHERE id = $1", id).Scan(&userWithDetails.ID, &userWithDetails.Username, &userWithDetails.CreatedAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching user: %v", err), http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT id, user_id, first_name, second_name, email, phone_number, position FROM user_details WHERE user_id = $1", id).Scan(
		&userWithDetails.Details.ID, &userWithDetails.Details.UserID, &userWithDetails.Details.FirstName, &userWithDetails.Details.SecondName, &userWithDetails.Details.Email, &userWithDetails.Details.PhoneNumber, &userWithDetails.Details.Position)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching user details: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userWithDetails)
}

// Удаление пользователя и его деталей
func DeleteUserWithDetails(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Начало транзакции
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Удаление деталей пользователя
	_, err = tx.Exec("DELETE FROM user_details WHERE user_id = $1", id)
	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error deleting user details: %v", err), http.StatusInternalServerError)
		return
	}

	// Удаление пользователя
	_, err = tx.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error deleting user: %v", err), http.StatusInternalServerError)
		return
	}

	// Подтверждение транзакции
	err = tx.Commit()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error committing transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Получение списка всех пользователей с их деталями
func GetAllUsersWithDetails(w http.ResponseWriter, r *http.Request) {
	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Выполнение запроса для получения всех пользователей и их деталей
	rows, err := db.Query(`
		SELECT u.id, u.username, u.created_at, ud.id, ud.user_id, ud.first_name, ud.second_name, ud.email, ud.phone_number, ud.position
		FROM users u
		JOIN user_details ud ON u.id = ud.user_id
	`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching users: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var usersWithDetails []UserWithDetails
	for rows.Next() {
		var userWithDetails UserWithDetails
		err := rows.Scan(
			&userWithDetails.ID, &userWithDetails.Username, &userWithDetails.CreatedAt,
			&userWithDetails.Details.ID, &userWithDetails.Details.UserID, &userWithDetails.Details.FirstName,
			&userWithDetails.Details.SecondName, &userWithDetails.Details.Email, &userWithDetails.Details.PhoneNumber,
			&userWithDetails.Details.Position,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}
		usersWithDetails = append(usersWithDetails, userWithDetails)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usersWithDetails)
}

func FetchAddressByID(db *sql.DB, addressID int) (Address, error) {
	var address Address
	err := db.QueryRow(
		"SELECT id, street, city, state, postal_code, country FROM addresses WHERE id = $1",
		addressID,
	).Scan(&address.ID, &address.Street, &address.City, &address.State, &address.PostalCode, &address.Country)
	return address, err
}


// Получение списка аптек
func GetPharmacies(w http.ResponseWriter, r *http.Request) {
	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name, address_id FROM pharmacies")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching pharmacies: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pharmacies []Pharmacy
	for rows.Next() {
		var pharmacy Pharmacy
		var addressID int
		if err := rows.Scan(&pharmacy.ID, &pharmacy.Name, &addressID); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}

		// Fetch address details.
		address, err := FetchAddressByID(db, addressID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching address: %v", err), http.StatusInternalServerError)
			return
		}
		pharmacy.Address = address
		pharmacies = append(pharmacies, pharmacy)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pharmacies)
}

// Получение аптеки по ID
func GetPharmacyByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	db, err := ConnectToDB()
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

// Create a new pharmacy with an address.
func CreatePharmacy(w http.ResponseWriter, r *http.Request) {
	var pharmacy Pharmacy
	err := json.NewDecoder(r.Body).Decode(&pharmacy)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert address first.
	var addressID int
	err = db.QueryRow(
		"INSERT INTO addresses(street, city, state, postal_code, country) VALUES($1, $2, $3, $4, $5) RETURNING id",
		pharmacy.Address.Street, pharmacy.Address.City, pharmacy.Address.State, pharmacy.Address.PostalCode, pharmacy.Address.Country,
	).Scan(&addressID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting address: %v", err), http.StatusInternalServerError)
		return
	}

	// Insert pharmacy with the new address ID.
	err = db.QueryRow(
		"INSERT INTO pharmacies(name, address_id) VALUES($1, $2) RETURNING id",
		pharmacy.Name, addressID,
	).Scan(&pharmacy.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting pharmacy: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pharmacy)
}

// Обновление информации о аптеке
func UpdatePharmacy(w http.ResponseWriter, r *http.Request) {
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

	db, err := ConnectToDB()
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
func DeletePharmacy(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	db, err := ConnectToDB()
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
func GetMedicines(w http.ResponseWriter, r *http.Request) {
    db, err := ConnectToDB()
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

        // Извлечение pharmacy_ids для текущего лекарства
        pharmacyRows, err := db.Query("SELECT pharmacy_id FROM pharmacy_medicines WHERE medicine_id = $1", medicine.ID)
        if err != nil {
            http.Error(w, fmt.Sprintf("Error fetching pharmacies for medicine: %v", err), http.StatusInternalServerError)
            return
        }

        var pharmacyIDs []int
        for pharmacyRows.Next() {
            var pharmacyID int
            if err := pharmacyRows.Scan(&pharmacyID); err != nil {
                http.Error(w, fmt.Sprintf("Error scanning pharmacy_id: %v", err), http.StatusInternalServerError)
                return
            }
            pharmacyIDs = append(pharmacyIDs, pharmacyID)
        }
        pharmacyRows.Close()

        medicine.PharmacyIDs = pharmacyIDs
        medicines = append(medicines, medicine)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(medicines)
}

// Получение лекарства по ID
func GetMedicineByID(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    db, err := ConnectToDB()
    if err != nil {
        http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    var medicine Medicine
    err = db.QueryRow("SELECT id, name, manufacturer, production_date, packaging, price FROM medicines WHERE id = $1",
        id).Scan(&medicine.ID, &medicine.Name, &medicine.Manufacturer, &medicine.ProductionDate, &medicine.Packaging, &medicine.Price)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error fetching medicine: %v", err), http.StatusInternalServerError)
        return
    }

    // Извлечение pharmacy_ids для текущего лекарства
    pharmacyRows, err := db.Query("SELECT pharmacy_id FROM pharmacy_medicines WHERE medicine_id = $1", medicine.ID)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error fetching pharmacies for medicine: %v", err), http.StatusInternalServerError)
        return
    }

    var pharmacyIDs []int
    for pharmacyRows.Next() {
        var pharmacyID int
        if err := pharmacyRows.Scan(&pharmacyID); err != nil {
            http.Error(w, fmt.Sprintf("Error scanning pharmacy_id: %v", err), http.StatusInternalServerError)
            return
        }
        pharmacyIDs = append(pharmacyIDs, pharmacyID)
    }
    pharmacyRows.Close()

    medicine.PharmacyIDs = pharmacyIDs

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(medicine)
}


// Создание нового лекарства
func CreateMedicine(w http.ResponseWriter, r *http.Request) {
    var medicine Medicine
    err := json.NewDecoder(r.Body).Decode(&medicine)
    if err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    db, err := ConnectToDB()
    if err != nil {
        http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
        return
    }
    defer db.Close()

	// Проверка наличия pharmacyID в таблице pharmacies
    for _, pharmacyID := range medicine.PharmacyIDs {
        var exists bool
        err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pharmacies WHERE id = $1)", pharmacyID).Scan(&exists)
        if err != nil {
            http.Error(w, fmt.Sprintf("Error checking pharmacy existence: %v", err), http.StatusInternalServerError)
            return
        }
        if !exists {
            http.Error(w, fmt.Sprintf("Pharmacy with ID %d does not exist", pharmacyID), http.StatusBadRequest)
            return
        }
    }

    // Вставка лекарства в таблицу medicines
    err = db.QueryRow("INSERT INTO medicines(name, manufacturer, production_date, packaging, price) VALUES($1, $2, $3, $4, $5) RETURNING id",
        medicine.Name, medicine.Manufacturer, medicine.ProductionDate, medicine.Packaging, medicine.Price).Scan(&medicine.ID)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error inserting medicine: %v", err), http.StatusInternalServerError)
        return
    }

    // Добавление связей с аптеками
    for _, pharmacyID := range medicine.PharmacyIDs {
        _, err := db.Exec("INSERT INTO pharmacy_medicines(pharmacy_id, medicine_id) VALUES($1, $2)", pharmacyID, medicine.ID)
        if err != nil {
            http.Error(w, fmt.Sprintf("Error inserting pharmacy-medicine relation: %v", err), http.StatusInternalServerError)
            return
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(medicine)
}


// Обновление информации о лекарстве
func UpdateMedicine(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedMedicine Medicine
	err = json.NewDecoder(r.Body).Decode(&updatedMedicine)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE medicines SET name = $1, manufacturer = $2, production_date = $3, packaging = $4, price = $5 WHERE id = $6",
		updatedMedicine.Name, updatedMedicine.Manufacturer, updatedMedicine.ProductionDate, updatedMedicine.Packaging, updatedMedicine.Price, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating medicine: %v", err), http.StatusInternalServerError)
		return
	}

	updatedMedicine.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedMedicine)
}

// Удаление лекарства
func DeleteMedicine(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	db, err := ConnectToDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to DB: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM medicines WHERE id = $1", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting medicine: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}