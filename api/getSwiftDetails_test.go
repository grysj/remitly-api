package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grysj/remitly-api/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSwiftDetails(t *testing.T) {
	require.NoError(t, testRedis.FlushDB(testCtx).Err())

	testHQ := db.Bank{
		Swift:      "AKBKMTMTXXX",
		ISO2:       "MT",
		Name:       "AKBANK T.A.S.",
		Type:       "Bank",
		Address:    "PORTOMASO BUSINESS TOWER",
		Town:       "ST. JULIAN'S",
		Country:    "MALTA",
		Timezone:   "Europe/Malta",
		Headquater: true,
	}
	err := db.AddBankToRedis(testRedis, testHQ)
	require.NoError(t, err)

	testBranch := db.Bank{
		Swift:      "ALBPPLP1BMW",
		ISO2:       "PL",
		Name:       "ALIOR BANK SPOLKA AKCYJNA",
		Type:       "Branch",
		Address:    "WARSZAWA, MAZOWIECKIE",
		Town:       "WARSZAWA",
		Country:    "POLAND",
		Timezone:   "Europe/Warsaw",
		Headquater: false,
	}
	err = db.AddBankToRedis(testRedis, testBranch)
	require.NoError(t, err)

	tests := []struct {
		name           string
		swiftCode      string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Headquarters Bank (XXX suffix)",
			swiftCode:      "AKBKMTMTXXX",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response getSwiftDetailsRes
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, "PORTOMASO BUSINESS TOWER", response.Address)
				assert.Equal(t, "AKBANK T.A.S.", response.BankName)
				assert.Equal(t, "MT", response.CountryISO2)
				assert.Equal(t, "MALTA", response.CountryName)
				assert.True(t, response.Headquater, "Should be headquarters due to XXX suffix")
				assert.Equal(t, "AKBKMTMTXXX", response.Swift)
			},
		},
		{
			name:           "Branch Bank",
			swiftCode:      "ALBPPLP1BMW",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response getSwiftDetailsRes
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, "WARSZAWA, MAZOWIECKIE", response.Address)
				assert.Equal(t, "ALIOR BANK SPOLKA AKCYJNA", response.BankName)
				assert.Equal(t, "PL", response.CountryISO2)
				assert.Equal(t, "POLAND", response.CountryName)
				assert.False(t, response.Headquater, "Should not be headquarters as suffix is not XXX")
				assert.Equal(t, "ALBPPLP1BMW", response.Swift)
				assert.Nil(t, response.Branches, "Branch should not have sub-branches")
			},
		},
		{
			name:           "Non-existent Swift Code (with XXX)",
			swiftCode:      "TESTMT11XXX",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Bank not found")
			},
		},
		{
			name:           "Invalid Swift Code Format (too short)",
			swiftCode:      "TESTXX",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid Swift code format")
			},
		},
		{
			name:           "Empty Swift Code",
			swiftCode:      "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Missing Swift code")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/v1/swift-codes/" + tt.swiftCode
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			testServer.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}
