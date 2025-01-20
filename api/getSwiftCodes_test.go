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

func TestGetSwiftCodes(t *testing.T) {
	require.NoError(t, testRedis.FlushDB(testCtx).Err())

	chileHQ := db.Bank{
		Swift:      "BCHICLRMXXX",
		ISO2:       "CL",
		Name:       "BANCO CENTRAL DE CHILE",
		Type:       "Bank",
		Address:    "",
		Town:       "SANTIAGO",
		Country:    "CHILE",
		Timezone:   "Pacific/Easter",
		Headquater: true,
	}
	err := db.AddBankToRedis(testRedis, chileHQ)
	require.NoError(t, err)

	chileBranches := []db.Bank{
		{
			Swift:      "BCHICLR10R3",
			ISO2:       "CL",
			Name:       "BANCO DE CHILE",
			Type:       "Branch",
			Address:    "21 DE MAYO 330 ARICA",
			Town:       "ARICA",
			Country:    "CHILE",
			Timezone:   "Pacific/Easter",
			Headquater: false,
		},
		{
			Swift:      "BCHICLR10R2",
			ISO2:       "CL",
			Name:       "BANCO DE CHILE",
			Type:       "Branch",
			Address:    "",
			Town:       "VINA DEL MAR",
			Country:    "CHILE",
			Timezone:   "Pacific/Easter",
			Headquater: false,
		},
	}

	for _, branch := range chileBranches {
		err := db.AddBankToRedis(testRedis, branch)
		require.NoError(t, err)
	}

	monacoBanks := []db.Bank{
		{
			Swift:      "BARCMCMXXXX",
			ISO2:       "MC",
			Name:       "BARCLAYS BANK PLC MONACO",
			Type:       "Bank",
			Address:    "31 AVENUE DE LA COSTA MONACO",
			Town:       "MONACO",
			Country:    "MONACO",
			Timezone:   "Europe/Monaco",
			Headquater: true,
		},
		{
			Swift:      "AGRIMCM1XXX",
			ISO2:       "MC",
			Name:       "CREDIT AGRICOLE MONACO",
			Type:       "Bank",
			Address:    "23 BOULEVARD PRINCESSE CHARLOTTE MONACO",
			Town:       "MONACO",
			Country:    "MONACO",
			Timezone:   "Europe/Monaco",
			Headquater: true,
		},
	}

	for _, bank := range monacoBanks {
		err := db.AddBankToRedis(testRedis, bank)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		countryCode    string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Country with Multiple Banks and Branches",
			countryCode:    "CL",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response getSwiftCodesRes
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, "CL", response.CountryISO2)
				assert.Equal(t, "CHILE", response.CountryName)
				assert.Len(t, response.SwiftCodes, 3)

				foundHQ := false
				for _, bank := range response.SwiftCodes {
					if bank.SwiftCode == "BCHICLRMXXX" {
						foundHQ = true
						assert.True(t, bank.IsHeadquater)
						assert.Equal(t, "BANCO CENTRAL DE CHILE", bank.BankName)
						assert.Equal(t, "", bank.Address)
					}
				}
				assert.True(t, foundHQ, "Should find headquarters bank")

				foundBranch := false
				for _, bank := range response.SwiftCodes {
					if bank.SwiftCode == "BCHICLR10R3" {
						foundBranch = true
						assert.False(t, bank.IsHeadquater)
						assert.Equal(t, "BANCO DE CHILE", bank.BankName)
						assert.Equal(t, "21 DE MAYO 330 ARICA", bank.Address)
					}
				}
				assert.True(t, foundBranch, "Should find branch with address")
			},
		},
		{
			name:           "Country with Multiple Independent Banks",
			countryCode:    "MC",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response getSwiftCodesRes
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, "MC", response.CountryISO2)
				assert.Equal(t, "MONACO", response.CountryName)
				assert.Len(t, response.SwiftCodes, 2)

				bankMap := make(map[string]BankInfo)
				for _, bank := range response.SwiftCodes {
					bankMap[bank.SwiftCode] = bank
					assert.True(t, bank.IsHeadquater)
				}

				barclays, exists := bankMap["BARCMCMXXXX"]
				assert.True(t, exists, "Should find Barclays bank")
				assert.Equal(t, "BARCLAYS BANK PLC MONACO", barclays.BankName)
				assert.Equal(t, "31 AVENUE DE LA COSTA MONACO", barclays.Address)

				agricole, exists := bankMap["AGRIMCM1XXX"]
				assert.True(t, exists, "Should find Credit Agricole bank")
				assert.Equal(t, "CREDIT AGRICOLE MONACO", agricole.BankName)
				assert.Equal(t, "23 BOULEVARD PRINCESSE CHARLOTTE MONACO", agricole.Address)
			},
		},
		{
			name:           "Non-existent Country",
			countryCode:    "XX",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response getSwiftCodesRes
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, "XX", response.CountryISO2)
				assert.Empty(t, response.CountryName)
				assert.Empty(t, response.SwiftCodes)
			},
		},
		{
			name:           "Empty Country Code",
			countryCode:    "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Missing country code")
			},
		},
		{
			name:           "Invalid Country Code Format",
			countryCode:    "USA",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid country code format")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/v1/swift-codes/country/" + tt.countryCode
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			testServer.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}
