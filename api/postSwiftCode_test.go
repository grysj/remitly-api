package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostSwiftCode(t *testing.T) {
	require.NoError(t, testServer.store.CleanDB(testCtx))
	tests := []struct {
		name           string
		requestBody    postSwiftCodeReq
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
		checkRedis     func(*testing.T)
	}{
		{
			name: "Successful Bank Addition",
			requestBody: postSwiftCodeReq{
				SwiftCode:   "EXAMPLEMCXXX",
				BankName:    "Example Bank",
				CountryISO2: "MC",
				CountryName: "Monaco",
				Address:     "123 Example Street, Monaco",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Message string `json:"message"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, "Bank added successfully", response.Message)
			},
			checkRedis: func(t *testing.T) {
				banks, err := testServer.store.GetBanksByISO2("MC")
				require.NoError(t, err)
				found := false
				for _, bank := range banks {
					if bank.Swift == "EXAMPLEMCXXX" {
						found = true
						assert.Equal(t, "EXAMPLE BANK", bank.Name)
						assert.Equal(t, "MC", bank.ISO2)
						assert.Equal(t, "123 Example Street, Monaco", bank.Address)
					}
				}
				assert.True(t, found)
			},
		},
		{
			name: "Missing Swift Code",
			requestBody: postSwiftCodeReq{
				BankName:    "Example Bank",
				CountryISO2: "MC",
				CountryName: "Monaco",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Swift code is required")
			},
			checkRedis: func(t *testing.T) {
				banks, err := testServer.store.GetBanksByISO2("MC")
				require.NoError(t, err)
				assert.Len(t, banks, 0)
			},
		},
		{
			name: "Invalid Country ISO2 Code",
			requestBody: postSwiftCodeReq{
				SwiftCode:   "EXAMPLEMCXXX",
				BankName:    "Example Bank",
				CountryISO2: "MONACO",
				CountryName: "Monaco",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid country ISO2 code")
			},
			checkRedis: func(t *testing.T) {
				banks, err := testServer.store.GetBanksByISO2("MONACO")
				require.NoError(t, err)
				assert.Len(t, banks, 0)
			},
		},
		{
			name: "Case Insensitive Input",
			requestBody: postSwiftCodeReq{
				SwiftCode:   "EXAMPLEMCXXX",
				BankName:    "Example Bank",
				CountryISO2: "mc",
				CountryName: "Monaco",
				Address:     "123 Example Street, Monaco",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Message string `json:"message"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, "Bank added successfully", response.Message)
			},
			checkRedis: func(t *testing.T) {
				banks, err := testServer.store.GetBanksByISO2("MC")
				require.NoError(t, err)
				found := false
				for _, bank := range banks {
					if bank.Swift == "EXAMPLEMCXXX" {
						found = true
						assert.Equal(t, "EXAMPLE BANK", bank.Name)
						assert.Equal(t, "MC", bank.ISO2)
					}
				}
				assert.True(t, found)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testServer.store.CleanDB(testCtx))

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/swift-codes", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+password)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			testServer.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			tt.checkRedis(t)
		})
	}
}
