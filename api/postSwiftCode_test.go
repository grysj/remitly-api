package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grysj/remitly-api/db"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostSwiftCode(t *testing.T) {
	require.NoError(t, testRedis.FlushDB(testCtx).Err())

	tests := []struct {
		name           string
		requestBody    postSwiftCodeReq
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
		checkRedis     func(*testing.T, *redis.Client)
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
			checkRedis: func(t *testing.T, client *redis.Client) {
				banks, err := db.GetBanksByISO2(client, "MC")
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
				assert.True(t, found, "Bank should be added to Redis")
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
			checkRedis: func(t *testing.T, client *redis.Client) {
				banks, err := db.GetBanksByISO2(client, "MC")
				require.NoError(t, err)
				assert.Len(t, banks, 0, "No banks should be added")
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
			checkRedis: func(t *testing.T, client *redis.Client) {
				banks, err := db.GetBanksByISO2(client, "MONACO")
				require.NoError(t, err)
				assert.Len(t, banks, 0, "No banks should be added")
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
			checkRedis: func(t *testing.T, client *redis.Client) {
				banks, err := db.GetBanksByISO2(client, "MC")
				require.NoError(t, err)

				found := false
				for _, bank := range banks {
					if bank.Swift == "EXAMPLEMCXXX" {
						found = true
						assert.Equal(t, "EXAMPLE BANK", bank.Name)
						assert.Equal(t, "MC", bank.ISO2)
					}
				}
				assert.True(t, found, "Bank should be added to Redis")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testRedis.FlushDB(testCtx).Err())

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			path := "/v1/swift-codes"
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			testServer.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			tt.checkRedis(t, testRedis)
		})
	}
}
