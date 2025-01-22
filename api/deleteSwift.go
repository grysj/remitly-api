package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/grysj/remitly-api/db"
	"github.com/grysj/remitly-api/util"
)

func (server *Server) deleteSwift(w http.ResponseWriter, r *http.Request) {
	swiftCode := r.PathValue("swiftcode")
	if swiftCode == "" {
		http.Error(w, "Missing SWIFT code", http.StatusBadRequest)
		return
	}

	var deleteErr error
	if util.CheckIfHeadquater(swiftCode) {
		prefix := strings.TrimSuffix(swiftCode, "XXX")
		if len(prefix) != 8 {
			http.Error(w, "Failed to delete bank", http.StatusBadRequest)
			return
		}
		deleteErr = server.store.DeleteBanksBySwiftPrefix(prefix)
	} else {
		deleteErr = server.store.DeleteBankFromDB(db.DeleteBankParams{
			Swift: swiftCode,
		})
	}

	if deleteErr != nil {
		http.Error(w, "Failed to delete bank", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully deleted",
	})
}
