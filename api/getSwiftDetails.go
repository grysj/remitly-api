package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/grysj/remitly-api/db"
	"github.com/grysj/remitly-api/util"
)

func (server *Server) getSwiftDetails(w http.ResponseWriter, r *http.Request) {
	swiftCode := r.PathValue("swiftcode")

	if swiftCode == "" {
		http.Error(w, "Missing Swift code", http.StatusBadRequest)
		return
	}

	if len(swiftCode) != 11 {
		http.Error(w, "Invalid Swift code format", http.StatusBadRequest)
		return
	}

	bank, err := db.GetBank(server.store, swiftCode)
	if err != nil {
		log.Printf("Error retrieving bank details: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if bank == nil {
		http.Error(w, "Bank not found", http.StatusNotFound)
		return
	}
	var response getSwiftDetailsRes

	response = getSwiftDetailsRes{
		Address:     bank.Address,
		BankName:    bank.Name,
		CountryISO2: bank.ISO2,
		CountryName: bank.Country,
		Headquater:  bank.Headquater,
		Swift:       bank.Swift,
	}

	if util.CheckIfHeadquater(swiftCode) {
		branches, err := db.GetBankBranches(server.store, swiftCode)
		if err != nil {
			log.Printf("Error retrieving bank branches: %v", err)
		}
		response.Branches = branches
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error generating response", http.StatusInternalServerError)
		return
	}

}

type getSwiftDetailsRes struct {
	Address     string                        `json:"address"`
	BankName    string                        `json:"bankName,omitempty"`
	CountryISO2 string                        `json:"countryISO2"`
	CountryName string                        `json:"countryName"`
	Headquater  bool                          `json:"isHeadquater"`
	Swift       string                        `json:"swiftCode"`
	Branches    []db.GetBranchesBySwiftResult `json:"branches,omitempty"`
}
