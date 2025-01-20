package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/grysj/remitly-api/db"
)

type BankInfo struct {
	Address      string `json:"address"`
	BankName     string `json:"bankName"`
	CountryISO2  string `json:"countryISO2"`
	IsHeadquater bool   `json:"isHeadquarter"`
	SwiftCode    string `json:"swiftCode"`
}

type getSwiftCodesRes struct {
	CountryISO2 string     `json:"countryISO2"`
	CountryName string     `json:"countryName"`
	SwiftCodes  []BankInfo `json:"swiftCodes"`
}

func (server *Server) getSwiftCodes(w http.ResponseWriter, r *http.Request) {
	countryCode := r.PathValue("countryISO2code")
	if countryCode == "" {
		http.Error(w, "Missing country code", http.StatusBadRequest)
		return
	}

	if len(countryCode) != 2 {
		http.Error(w, "Invalid country code format", http.StatusBadRequest)
		return
	}

	countryCode = strings.ToUpper(countryCode)
	response := getSwiftCodesRes{
		CountryISO2: countryCode,
		SwiftCodes:  make([]BankInfo, 0),
	}

	countryName, err := db.GetCountryNameByISO2(server.store, countryCode)
	if err != nil {
		log.Printf("Error retrieving country name for %s: %v", countryCode, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response.CountryName = countryName

	banks, err := db.GetBanksByISO2(server.store, countryCode)
	if err != nil {
		log.Printf("Error retrieving banks for country %s: %v", countryCode, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, bank := range banks {
		bankInfo := BankInfo{
			Address:      bank.Address,
			BankName:     bank.Name,
			CountryISO2:  bank.ISO2,
			IsHeadquater: bank.Headquater,
			SwiftCode:    bank.Swift,
		}
		response.SwiftCodes = append(response.SwiftCodes, bankInfo)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error generating response", http.StatusInternalServerError)
		return
	}
}
