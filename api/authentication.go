package api

import (
	"net/http"
	"strings"
)
func Middleware(password string, endpoint http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) != 2 || strings.ToLower(fields[0]) != "bearer" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if fields[1] != password {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		endpoint(w, r)
	}
}
