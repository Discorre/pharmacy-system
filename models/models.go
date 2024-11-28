package models

// Pharmacy представляет информацию об аптеке
type Pharmacy struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// Medicine представляет информацию о лекарстве
type Medicine struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Manufacturer   string  `json:"manufacturer"`
	ProductionDate string  `json:"production_date"`
	Packaging      string  `json:"packaging"`
	Price          float64 `json:"price"`
	PharmacyIDs    []int   `json:"pharmacy_ids"`
}
