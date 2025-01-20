package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grysj/remitly-api/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteSwift(t *testing.T) {
	require.NoError(t, testRedis.FlushDB(testCtx).Err())

	testBanks := []db.Bank{
		{
			Swift:      "BCHICLRMXXX",
			ISO2:       "CL",
			Name:       "BANCO CENTRAL DE CHILE",
			Headquater: true,
			Country:    "CHILE",
		},
		{
			Swift:      "BCHICLRM001",
			ISO2:       "CL",
			Name:       "BANCO DE CHILE BRANCH 1",
			Headquater: false,
			Country:    "CHILE",
		},
		{
			Swift:      "BCHICLRM002",
			ISO2:       "CL",
			Name:       "BANCO DE CHILE BRANCH 2",
			Headquater: false,
			Country:    "CHILE",
		},
	}

	for _, bank := range testBanks {
		err := db.AddBankToRedis(testRedis, bank)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		swiftCode      string
		expectedStatus int
	}{
		{
			name:           "Delete Headquarters Successfully",
			swiftCode:      "BCHICLRMXXX",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Delete Single Branch Successfully",
			swiftCode:      "BCHICLRM001",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty Swift Code",
			swiftCode:      "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Nonexistent Swift Code",
			swiftCode:      "XXXXXXXXXX",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/v1/swift-codes/" + tt.swiftCode
			req := httptest.NewRequest(http.MethodDelete, path, nil)
			w := httptest.NewRecorder()

			testServer.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
