package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/grysj/remitly-api/db"
)

type postSwiftCodeReq struct {
	Address     string `json:"address"`
	BankName    string `json:"bankName"`
	CountryISO2 string `json:"countryISO2"`
	CountryName string `json:"countryName"`
	SwiftCode   string `json:"swiftCode"`
}

func (server *Server) postSwiftCode(w http.ResponseWriter, r *http.Request) {
	var newBank postSwiftCodeReq
	if err := json.NewDecoder(r.Body).Decode(&newBank); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if newBank.SwiftCode == "" {
		http.Error(w, "Swift code is required", http.StatusBadRequest)
		return
	}

	if len(newBank.CountryISO2) != 2 {
		http.Error(w, "Invalid country ISO2 code", http.StatusBadRequest)
		return
	}

	bankToAdd := db.Bank{
		Swift:   newBank.SwiftCode,
		ISO2:    strings.ToUpper(newBank.CountryISO2),
		Name:    strings.ToUpper(newBank.BankName),
		Address: newBank.Address,
		Country: strings.ToUpper(newBank.CountryName),
	}

	err := db.AddBankToRedis(server.store, bankToAdd)
	if err != nil {
		log.Printf("Error adding bank: %v", err)
		http.Error(w, "Failed to add bank", http.StatusInternalServerError)
		return
	}

	response := struct {
		Message string `json:"message"`
	}{
		Message: "Bank added successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error generating response", http.StatusInternalServerError)
		return
	}
}
