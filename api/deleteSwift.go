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
		deleteErr = db.DeleteBanksBySwiftPrefix(server.store, prefix)
	} else {
		deleteErr = db.DeleteBankFromRedis(server.store, db.DeleteBankParams{
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
