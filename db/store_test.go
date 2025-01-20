package db

import (
	"sort"
	"testing"

	"github.com/grysj/remitly-api/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddBanksToRedis(t *testing.T) {
	tests := []struct {
		name    string
		rows    []parser.CsvRow
		wantErr bool
		verify  func(*testing.T)
	}{
		{
			name: "successful_import",
			rows: []parser.CsvRow{
				{
					ISO2:     "AL",
					Swift:    "AAISALTRXXX",
					Type:     "BIC11",
					Name:     "UNITED BANK OF ALBANIA SH.A",
					Address:  "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023",
					Town:     "TIRANA",
					Country:  "ALBANIA",
					Timezone: "Europe/Tirane",
				},
				{
					ISO2:     "MC",
					Swift:    "BAERMCMCXXX",
					Type:     "BIC11",
					Name:     "BANK JULIUS BAER (MONACO) S.A.M.",
					Address:  "12 BOULEVARD DES MOULINS  MONACO, MONACO, 98000",
					Town:     "MONACO",
					Country:  "MONACO",
					Timezone: "Europe/Monaco",
				},
			},
			wantErr: false,
			verify: func(t *testing.T) {
				result, err := testRedis.HGetAll(testCtx, "swiftCode:AAISALTRXXX").Result()
				require.NoError(t, err)
				require.NotEmpty(t, result)

				var bankData GetBankByIsoResult
				err = testRedis.HGetAll(testCtx, "swiftCode:AAISALTRXXX").Scan(&bankData)
				require.NoError(t, err)
				assert.Equal(t, "AL", bankData.ISO2)
				assert.Equal(t, "UNITED BANK OF ALBANIA SH.A", bankData.Name)

				members, err := testRedis.SMembers(testCtx, "idx:countryISO2:AL").Result()
				require.NoError(t, err)
				assert.Contains(t, members, "swiftCode:AAISALTRXXX")

				var bankData2 GetBankByIsoResult
				err = testRedis.HGetAll(testCtx, "swiftCode:BAERMCMCXXX").Scan(&bankData2)
				require.NoError(t, err)
				assert.Equal(t, "MC", bankData2.ISO2)
				assert.Equal(t, "BANK JULIUS BAER (MONACO) S.A.M.", bankData2.Name)
			},
		},
		{
			name: "multiple_banks_same_country",
			rows: []parser.CsvRow{
				{
					ISO2:     "BG",
					Swift:    "ABIEBGS1XXX",
					Type:     "BIC11",
					Name:     "ABV INVESTMENTS LTD",
					Address:  "TSAR ASEN 20  VARNA, VARNA, 9002",
					Town:     "VARNA",
					Country:  "BULGARIA",
					Timezone: "Europe/Sofia",
				},
				{
					ISO2:     "BG",
					Swift:    "ADCRBGS1XXX",
					Type:     "BIC11",
					Name:     "ADAMANT CAPITAL PARTNERS AD",
					Address:  "JAMES BOURCHIER BLVD 76A HILL TOWER SOFIA, SOFIA, 1421",
					Town:     "SOFIA",
					Country:  "BULGARIA",
					Timezone: "Europe/Sofia",
				},
			},
			wantErr: false,
			verify: func(t *testing.T) {

				result1, err := testRedis.HGetAll(testCtx, "swiftCode:ABIEBGS1XXX").Result()
				require.NoError(t, err)
				require.NotEmpty(t, result1)

				result2, err := testRedis.HGetAll(testCtx, "swiftCode:ADCRBGS1XXX").Result()
				require.NoError(t, err)
				require.NotEmpty(t, result2)

				members, err := testRedis.SMembers(testCtx, "idx:countryISO2:BG").Result()
				require.NoError(t, err)
				assert.Len(t, members, 2)
				assert.Contains(t, members, "swiftCode:ABIEBGS1XXX")
				assert.Contains(t, members, "swiftCode:ADCRBGS1XXX")

				var bank1 GetBankByIsoResult
				err = testRedis.HGetAll(testCtx, "swiftCode:ABIEBGS1XXX").Scan(&bank1)
				require.NoError(t, err)
				assert.Equal(t, "BG", bank1.ISO2)
				assert.Equal(t, "ABV INVESTMENTS LTD", bank1.Name)

				var bank2 GetBankByIsoResult
				err = testRedis.HGetAll(testCtx, "swiftCode:ADCRBGS1XXX").Scan(&bank2)
				require.NoError(t, err)
				assert.Equal(t, "BG", bank2.ISO2)
				assert.Equal(t, "ADAMANT CAPITAL PARTNERS AD", bank2.Name)
			},
		},
		{
			name:    "empty_input",
			rows:    []parser.CsvRow{},
			wantErr: false,
			verify: func(t *testing.T) {
				keys, err := testRedis.Keys(testCtx, "swiftCode:*").Result()
				require.NoError(t, err)
				assert.Empty(t, keys)

				keys, err = testRedis.Keys(testCtx, "idx:countryISO2:*").Result()
				require.NoError(t, err)
				assert.Empty(t, keys)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRedis.FlushDB(testCtx).Err()
			require.NoError(t, err)

			err = AddBanksToRedis(testRedis, tt.rows)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.verify != nil {
				tt.verify(t)
			}
		})
	}
}

func TestAddBankToRedis(t *testing.T) {
	tests := []struct {
		name    string
		bank    Bank
		wantErr bool
		verify  func(*testing.T, Bank)
	}{
		{
			name: "successful_add",
			bank: Bank{
				ISO2:     "AL",
				Swift:    "AAISALTRXXX",
				Type:     "BIC11",
				Name:     "United Bank of Albania SH.A",
				Address:  "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023",
				Town:     "TIRANA",
				Country:  "ALBANIA",
				Timezone: "Europe/Tirane",
			},
			wantErr: false,
			verify: func(t *testing.T, _ Bank) {
				bankData := &Bank{}
				err := testRedis.HGetAll(testCtx, "swiftCode:AAISALTRXXX").Scan(bankData)
				require.NoError(t, err)
				assert.Equal(t, "AL", bankData.ISO2)
				assert.Equal(t, "UNITED BANK OF ALBANIA SH.A", bankData.Name)
				keys, err := testRedis.SMembers(testCtx, "idx:countryISO2:AL").Result()
				require.NoError(t, err)
				assert.Contains(t, keys, "swiftCode:AAISALTRXXX")
			},
		},
		{
			name: "add_with_lowercase_input",
			bank: Bank{
				ISO2:     "bg",
				Swift:    "ABIEBGS1XXX",
				Type:     "BIC11",
				Name:     "ABV Investments Ltd",
				Address:  "TSAR ASEN 20  VARNA, VARNA, 9002",
				Town:     "VARNA",
				Country:  "BULGARIA",
				Timezone: "Europe/Sofia",
			},
			wantErr: false,
			verify: func(t *testing.T, _ Bank) {
				bankData := &Bank{}
				err := testRedis.HGetAll(testCtx, "swiftCode:ABIEBGS1XXX").Scan(bankData)
				require.NoError(t, err)
				assert.Equal(t, "BG", bankData.ISO2)
				assert.Equal(t, "ABV INVESTMENTS LTD", bankData.Name)

				keys, err := testRedis.SMembers(testCtx, "idx:countryISO2:BG").Result()
				require.NoError(t, err)
				assert.Contains(t, keys, "swiftCode:ABIEBGS1XXX")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRedis.FlushDB(testCtx).Err()
			require.NoError(t, err)

			err = AddBankToRedis(testRedis, tt.bank)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.verify != nil {
				tt.verify(t, tt.bank)
			}
		})
	}
}

func TestDeleteBankFromRedis(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		bank    DeleteBankParams
		wantErr bool
		verify  func(*testing.T)
	}{
		{
			name: "successful_delete",
			setup: func() {
				bank := Bank{
					ISO2:     "AL",
					Swift:    "AAISALTRXXX",
					Type:     "BIC11",
					Name:     "UNITED BANK OF ALBANIA SH.A",
					Address:  "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023",
					Town:     "TIRANA",
					Country:  "ALBANIA",
					Timezone: "Europe/Tirane",
				}
				err := AddBankToRedis(testRedis, bank)
				require.NoError(t, err)
			},
			bank: DeleteBankParams{
				Swift: "AAISALTRXXX",
			},
			wantErr: false,
			verify: func(t *testing.T) {
				exists, err := testRedis.Exists(testCtx, "swiftCode:AAISALTRXXX").Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), exists)

				keys, err := testRedis.SMembers(testCtx, "idx:countryISO2:AL").Result()
				require.NoError(t, err)
				assert.NotContains(t, keys, "swiftCode:AAISALTRXXX")
			},
		},
		{
			name: "delete_with_multiple_banks_in_country",
			setup: func() {
				banks := []Bank{
					{
						ISO2:     "BG",
						Swift:    "ABIEBGS1XXX",
						Type:     "BIC11",
						Name:     "ABV INVESTMENTS LTD",
						Address:  "TSAR ASEN 20  VARNA, VARNA, 9002",
						Town:     "VARNA",
						Country:  "BULGARIA",
						Timezone: "Europe/Sofia",
					},
					{
						ISO2:     "BG",
						Swift:    "ADCRBGS1XXX",
						Type:     "BIC11",
						Name:     "ADAMANT CAPITAL PARTNERS AD",
						Address:  "JAMES BOURCHIER BLVD 76A HILL TOWER SOFIA, SOFIA, 1421",
						Town:     "SOFIA",
						Country:  "BULGARIA",
						Timezone: "Europe/Sofia",
					},
				}
				for _, bank := range banks {
					err := AddBankToRedis(testRedis, bank)
					require.NoError(t, err)
				}
			},
			bank: DeleteBankParams{
				Swift: "ABIEBGS1XXX",
			},
			wantErr: false,
			verify: func(t *testing.T) {
				exists, err := testRedis.Exists(testCtx, "swiftCode:ABIEBGS1XXX").Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), exists)

				exists, err = testRedis.Exists(testCtx, "swiftCode:ADCRBGS1XXX").Result()
				require.NoError(t, err)
				assert.Equal(t, int64(1), exists)

				keys, err := testRedis.SMembers(testCtx, "idx:countryISO2:BG").Result()
				require.NoError(t, err)
				assert.NotContains(t, keys, "swiftCode:ABIEBGS1XXX")
				assert.Contains(t, keys, "swiftCode:ADCRBGS1XXX")
			},
		},
		{
			name: "delete_nonexistent_bank",
			bank: DeleteBankParams{
				Swift: "NONEXISTXXX",
			},
			wantErr: false,
			verify: func(t *testing.T) {
				err := DeleteBankFromRedis(testRedis, DeleteBankParams{Swift: "NONEXISTXXX"})
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRedis.FlushDB(testCtx).Err()
			require.NoError(t, err)

			if tt.setup != nil {
				tt.setup()
			}

			err = DeleteBankFromRedis(testRedis, tt.bank)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.verify != nil {
				tt.verify(t)
			}
		})
	}
}

func TestGetBanksByISO2(t *testing.T) {
	if err := testRedis.FlushDB(testCtx).Err(); err != nil {
		t.Fatalf("Failed to flush database: %v", err)
	}

	testBanks := []struct {
		data Bank
		key  string
	}{
		{
			data: Bank{
				Swift:      "BCGMMCM1XXX",
				ISO2:       "MC",
				Name:       "BNP PARIBAS WEALTH MANAGEMENT MONACO",
				Type:       "BIC11",
				Address:    "27 BOULEVARD PRINCESSE CHARLOTTE  MONACO, MONACO, 98000",
				Town:       "MONACO",
				Country:    "MONACO",
				Timezone:   "Europe/Monaco",
				Headquater: true,
			},
			key: "swiftCode:BCGMMCM1XXX",
		},
		{
			data: Bank{
				Swift:      "BARCMCC1XXX",
				ISO2:       "MC",
				Name:       "BARCLAYS BANK S.A",
				Type:       "BIC11",
				Address:    "31 AVENUE DE LA COSTA  MONACO, MONACO, 98000",
				Town:       "MONACO",
				Country:    "MONACO",
				Timezone:   "Europe/Monaco",
				Headquater: false,
			},
			key: "swiftCode:BARCMCC1XXX",
		},
	}

	for _, bank := range testBanks {
		bankData := map[string]interface{}{
			"swiftCode":    bank.data.Swift,
			"countryISO2":  bank.data.ISO2,
			"bankName":     bank.data.Name,
			"address":      bank.data.Address,
			"isHeadquater": bank.data.Headquater,
		}

		if err := testRedis.HSet(testCtx, bank.key, bankData).Err(); err != nil {
			t.Fatalf("Failed to add test bank: %v", err)
		}
		if err := testRedis.SAdd(testCtx, iso2IndexKey+":"+bank.data.ISO2, bank.key).Err(); err != nil {
			t.Fatalf("Failed to add bank to ISO2 index: %v", err)
		}
	}
	tests := []struct {
		name     string
		iso2     string
		wantLen  int
		wantBank *GetBankByIsoResult
	}{
		{
			name:    "Valid ISO2 - MC",
			iso2:    "MC",
			wantLen: 2,
			wantBank: &GetBankByIsoResult{
				Swift:      "BCGMMCM1XXX",
				ISO2:       "MC",
				Name:       "BNP PARIBAS WEALTH MANAGEMENT MONACO",
				Address:    "27 BOULEVARD PRINCESSE CHARLOTTE  MONACO, MONACO, 98000",
				Headquater: true,
			},
		},
		{
			name:    "Non-existent ISO2",
			iso2:    "XX",
			wantLen: 0,
		},
		{
			name:    "Empty ISO2",
			iso2:    "",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBanksByISO2(testRedis, tt.iso2)
			if err != nil {
				t.Errorf("GetBanksByISO2() error = %v", err)
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf("GetBanksByISO2() got %d banks, want %d", len(got), tt.wantLen)
				return
			}

			if tt.wantBank != nil && len(got) > 0 {
				found := false
				for _, bank := range got {
					if bank.Swift == tt.wantBank.Swift {
						found = true
						if bank.ISO2 != tt.wantBank.ISO2 {
							t.Errorf("Bank ISO2 = %v, want %v", bank.ISO2, tt.wantBank.ISO2)
						}
						if bank.Name != tt.wantBank.Name {
							t.Errorf("Bank Name = %v, want %v", bank.Name, tt.wantBank.Name)
						}
						if bank.Address != tt.wantBank.Address {
							t.Errorf("Bank Address = %v, want %v", bank.Address, tt.wantBank.Address)
						}
						if bank.Headquater != tt.wantBank.Headquater {
							t.Errorf("Bank Headquater = %v, want %v", bank.Headquater, tt.wantBank.Headquater)
						}
						break
					}
				}
				if !found {
					t.Errorf("Expected bank with Swift %v not found", tt.wantBank.Swift)
				}
			}
		})
	}
}
func TestGetBankBranches(t *testing.T) {
	tests := []struct {
		name    string
		swift   string
		setup   func(*testing.T)
		want    []GetBranchesBySwiftResult
		wantErr bool
	}{
		{
			name:  "get_branches",
			swift: "ALBPPLPWXXX",
			setup: func(t *testing.T) {
				headOffice := Bank{
					ISO2:     "PL",
					Swift:    "ALBPPLPWXXX",
					Type:     "BIC11",
					Name:     "ALIOR BANK HEAD OFFICE",
					Address:  "LOPUSZANSKA BUSINESS PARK",
					Town:     "WARSZAWA",
					Country:  "POLAND",
					Timezone: "Europe/Warsaw",
				}
				err := AddBankToRedis(testRedis, headOffice)
				require.NoError(t, err)

				branches := []Bank{
					{
						ISO2:     "PL",
						Swift:    "ALBPPLPW001",
						Type:     "BIC11",
						Name:     "ALIOR BANK BRANCH 1",
						Address:  "Branch Address 1",
						Town:     "WARSZAWA",
						Country:  "POLAND",
						Timezone: "Europe/Warsaw",
					},
					{
						ISO2:     "PL",
						Swift:    "ALBPPLPW002",
						Type:     "BIC11",
						Name:     "ALIOR BANK BRANCH 2",
						Address:  "Branch Address 2",
						Town:     "KRAKOW",
						Country:  "POLAND",
						Timezone: "Europe/Warsaw",
					},
				}
				for _, branch := range branches {
					err := AddBankToRedis(testRedis, branch)
					require.NoError(t, err)
				}

				members, err := testRedis.SMembers(testCtx, "branch:ALBPPLPW").Result()
				require.NoError(t, err)
				assert.Contains(t, members, "ALBPPLPW001")
				assert.Contains(t, members, "ALBPPLPW002")
			},
			want: []GetBranchesBySwiftResult{
				{
					ISO2:       "PL",
					Swift:      "ALBPPLPW001",
					Name:       "ALIOR BANK BRANCH 1",
					Address:    "Branch Address 1",
					Headquater: false,
				},
				{
					ISO2:       "PL",
					Swift:      "ALBPPLPW002",
					Name:       "ALIOR BANK BRANCH 2",
					Address:    "Branch Address 2",
					Headquater: false,
				},
			},
			wantErr: false,
		},
		{
			name:  "no_branches",
			swift: "AAISALTRXXX",
			setup: func(t *testing.T) {
				bank := Bank{
					ISO2:     "AL",
					Swift:    "AAISALTRXXX",
					Type:     "BIC11",
					Name:     "UNITED BANK OF ALBANIA",
					Address:  "Main Address",
					Town:     "TIRANA",
					Country:  "ALBANIA",
					Timezone: "Europe/Tirane",
				}
				err := AddBankToRedis(testRedis, bank)
				require.NoError(t, err)
			},
			want:    []GetBranchesBySwiftResult{},
			wantErr: false,
		},
		{
			name:    "nonexistent_bank",
			swift:   "NONEXISTXXX",
			setup:   nil,
			want:    []GetBranchesBySwiftResult{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRedis.FlushDB(testCtx).Err()
			require.NoError(t, err)

			if tt.setup != nil {
				tt.setup(t)
			}

			got, err := GetBankBranches(testRedis, tt.swift)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			sort.Slice(got, func(i, j int) bool {
				return got[i].Swift < got[j].Swift
			})
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].Swift < tt.want[j].Swift
			})

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetBank(t *testing.T) {
	tests := []struct {
		name    string
		swift   string
		setup   func(*testing.T)
		want    *GetBankBySwiftResult
		wantErr bool
	}{
		{
			name:  "successful_headquarter_bank_retrieval",
			swift: "AAISALTRXXX",
			setup: func(t *testing.T) {
				bank := Bank{
					ISO2:     "AL",
					Swift:    "AAISALTRXXX",
					Type:     "BIC11",
					Name:     "UNITED BANK OF ALBANIA SH.A",
					Address:  "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023",
					Town:     "TIRANA",
					Country:  "ALBANIA",
					Timezone: "Europe/Tirane",
				}
				err := AddBankToRedis(testRedis, bank)
				require.NoError(t, err)
			},
			want: &GetBankBySwiftResult{
				Swift:      "AAISALTRXXX",
				ISO2:       "AL",
				Name:       "UNITED BANK OF ALBANIA SH.A",
				Address:    "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023",
				Country:    "ALBANIA",
				Headquater: true,
			},
			wantErr: false,
		},
		{
			name:  "successful_branch_bank_retrieval",
			swift: "ALBPPLPW001",
			setup: func(t *testing.T) {
				headOffice := Bank{
					ISO2:     "PL",
					Swift:    "ALBPPLPWXXX",
					Type:     "BIC11",
					Name:     "ALIOR BANK HEAD OFFICE",
					Address:  "LOPUSZANSKA BUSINESS PARK",
					Town:     "WARSZAWA",
					Country:  "POLAND",
					Timezone: "Europe/Warsaw",
				}
				err := AddBankToRedis(testRedis, headOffice)
				require.NoError(t, err)

				branch := Bank{
					ISO2:     "PL",
					Swift:    "ALBPPLPW001",
					Type:     "BIC11",
					Name:     "ALIOR BANK BRANCH 1",
					Address:  "Branch Address 1",
					Town:     "WARSZAWA",
					Country:  "POLAND",
					Timezone: "Europe/Warsaw",
				}
				err = AddBankToRedis(testRedis, branch)
				require.NoError(t, err)
			},
			want: &GetBankBySwiftResult{
				Swift:      "ALBPPLPW001",
				ISO2:       "PL",
				Name:       "ALIOR BANK BRANCH 1",
				Address:    "Branch Address 1",
				Country:    "POLAND",
				Headquater: false,
			},
			wantErr: false,
		},
		{
			name:    "nonexistent_bank",
			swift:   "NONEXISTXXX",
			setup:   nil,
			want:    nil,
			wantErr: false,
		},
		{
			name:  "case_insensitive_swift_code",
			swift: "aaisaltrxxx",
			setup: func(t *testing.T) {
				bank := Bank{
					ISO2:     "AL",
					Swift:    "AAISALTRXXX",
					Type:     "BIC11",
					Name:     "UNITED BANK OF ALBANIA SH.A",
					Address:  "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023",
					Town:     "TIRANA",
					Country:  "ALBANIA",
					Timezone: "Europe/Tirane",
				}
				err := AddBankToRedis(testRedis, bank)
				require.NoError(t, err)
			},
			want: &GetBankBySwiftResult{
				Swift:      "AAISALTRXXX",
				ISO2:       "AL",
				Name:       "UNITED BANK OF ALBANIA SH.A",
				Address:    "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023",
				Country:    "ALBANIA",
				Headquater: true,
			},
			wantErr: false,
		},
		{
			name:  "multiple_banks_different_swift_codes",
			swift: "BAERMCMCXXX",
			setup: func(t *testing.T) {
				banks := []Bank{
					{
						ISO2:     "AL",
						Swift:    "AAISALTRXXX",
						Type:     "BIC11",
						Name:     "UNITED BANK OF ALBANIA SH.A",
						Address:  "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023",
						Town:     "TIRANA",
						Country:  "ALBANIA",
						Timezone: "Europe/Tirane",
					},
					{
						ISO2:     "MC",
						Swift:    "BAERMCMCXXX",
						Type:     "BIC11",
						Name:     "BANK JULIUS BAER (MONACO) S.A.M.",
						Address:  "12 BOULEVARD DES MOULINS  MONACO, MONACO, 98000",
						Town:     "MONACO",
						Country:  "MONACO",
						Timezone: "Europe/Monaco",
					},
				}
				for _, bank := range banks {
					err := AddBankToRedis(testRedis, bank)
					require.NoError(t, err)
				}
			},
			want: &GetBankBySwiftResult{
				Swift:      "BAERMCMCXXX",
				ISO2:       "MC",
				Name:       "BANK JULIUS BAER (MONACO) S.A.M.",
				Address:    "12 BOULEVARD DES MOULINS  MONACO, MONACO, 98000",
				Country:    "MONACO",
				Headquater: true,
			},
			wantErr: false,
		},
		{
			name:  "bank_with_optional_fields",
			swift: "TESTBANK1XXX",
			setup: func(t *testing.T) {
				bank := Bank{
					ISO2:     "US",
					Swift:    "TESTBANK1XXX",
					Type:     "BIC11",
					Name:     "TEST BANK",
					Address:  "",
					Town:     "",
					Country:  "UNITED STATES",
					Timezone: "",
				}
				err := AddBankToRedis(testRedis, bank)
				require.NoError(t, err)
			},
			want: &GetBankBySwiftResult{
				Swift:      "TESTBANK1XXX",
				ISO2:       "US",
				Name:       "TEST BANK",
				Address:    "",
				Country:    "UNITED STATES",
				Headquater: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRedis.FlushDB(testCtx).Err()
			require.NoError(t, err)

			if tt.setup != nil {
				tt.setup(t)
			}

			got, err := GetBank(testRedis, tt.swift)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetCountryNameByISO2(t *testing.T) {
	tests := []struct {
		name    string
		iso2    string
		setup   func(*testing.T)
		want    string
		wantErr bool
	}{
		{
			name: "successful_country_retrieval",
			iso2: "PL",
			setup: func(t *testing.T) {
				err := testRedis.HSet(testCtx, "countries", "PL", "POLAND").Err()
				require.NoError(t, err)
			},
			want:    "POLAND",
			wantErr: false,
		},
		{
			name: "case_insensitive_code",
			iso2: "mc",
			setup: func(t *testing.T) {
				err := testRedis.HSet(testCtx, "countries", "MC", "MONACO").Err()
				require.NoError(t, err)
			},
			want:    "MONACO",
			wantErr: false,
		},
		{
			name:    "nonexistent_country",
			iso2:    "XX",
			setup:   nil,
			want:    "",
			wantErr: false,
		},
		{
			name: "multiple_countries_in_db",
			iso2: "DE",
			setup: func(t *testing.T) {
				err := testRedis.HSet(testCtx, "countries",
					"DE", "GERMANY",
					"FR", "FRANCE",
					"IT", "ITALY",
				).Err()
				require.NoError(t, err)
			},
			want:    "GERMANY",
			wantErr: false,
		},
		{
			name:    "empty_country_code",
			iso2:    "",
			setup:   nil,
			want:    "",
			wantErr: false,
		},
		{
			name: "country_with_spaces",
			iso2: "US",
			setup: func(t *testing.T) {
				err := testRedis.HSet(testCtx, "countries", "US", "UNITED STATES").Err()
				require.NoError(t, err)
			},
			want:    "UNITED STATES",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRedis.FlushDB(testCtx).Err()
			require.NoError(t, err)

			if tt.setup != nil {
				tt.setup(t)
			}

			got, err := GetCountryNameByISO2(testRedis, tt.iso2)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteBanksBySwiftPrefix(t *testing.T) {
	tests := []struct {
		name        string
		swiftPrefix string
		setup       func(*testing.T)
		verify      func(*testing.T)
		wantErr     bool
	}{
		{
			name:        "delete_bank_with_branches",
			swiftPrefix: "BCHICLRM",
			setup: func(t *testing.T) {
				hq := Bank{
					Swift:    "BCHICLRMXXX",
					ISO2:     "CL",
					Name:     "BANCO DE CHILE",
					Type:     "BIC11",
					Address:  "CENTRAL ADDRESS",
					Town:     "SANTIAGO",
					Country:  "CHILE",
					Timezone: "Pacific/Easter",
				}
				err := AddBankToRedis(testRedis, hq)
				require.NoError(t, err)

				branches := []Bank{
					{
						Swift:    "BCHICLRM001",
						ISO2:     "CL",
						Name:     "BANCO DE CHILE BRANCH 1",
						Type:     "BIC11",
						Address:  "BRANCH 1 ADDRESS",
						Town:     "ARICA",
						Country:  "CHILE",
						Timezone: "Pacific/Easter",
					},
					{
						Swift:    "BCHICLRM002",
						ISO2:     "CL",
						Name:     "BANCO DE CHILE BRANCH 2",
						Type:     "BIC11",
						Address:  "BRANCH 2 ADDRESS",
						Town:     "VINA DEL MAR",
						Country:  "CHILE",
						Timezone: "Pacific/Easter",
					},
				}

				for _, branch := range branches {
					err := AddBankToRedis(testRedis, branch)
					require.NoError(t, err)
				}
			},
			verify: func(t *testing.T) {
				hqKey := bankKeyPrefix + "BCHICLRMXXX"
				exists, err := testRedis.Exists(testCtx, hqKey).Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), exists, "headquarters should be deleted")

				branchKeys := []string{
					bankKeyPrefix + "BCHICLRM001",
					bankKeyPrefix + "BCHICLRM002",
				}
				for _, key := range branchKeys {
					exists, err := testRedis.Exists(testCtx, key).Result()
					require.NoError(t, err)
					assert.Equal(t, int64(0), exists, "branch should be deleted: "+key)
				}

				branchSetKey := "branch:BCHICLRM"
				exists, err = testRedis.Exists(testCtx, branchSetKey).Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), exists, "branch set should be deleted")

				isoKey := iso2IndexKey + ":CL"
				members, err := testRedis.SMembers(testCtx, isoKey).Result()
				require.NoError(t, err)
				assert.NotContains(t, members, hqKey, "ISO2 index should not contain HQ")
				for _, branchKey := range branchKeys {
					assert.NotContains(t, members, branchKey, "ISO2 index should not contain branch")
				}
			},
			wantErr: false,
		},
		{
			name:        "delete_headquarters_only",
			swiftPrefix: "AGRIMCM1",
			setup: func(t *testing.T) {
				hq := Bank{
					Swift:    "AGRIMCM1XXX",
					ISO2:     "MC",
					Name:     "CREDIT AGRICOLE MONACO",
					Type:     "BIC11",
					Address:  "23 BOULEVARD PRINCESSE CHARLOTTE MONACO",
					Town:     "MONACO",
					Country:  "MONACO",
					Timezone: "Europe/Monaco",
				}
				err := AddBankToRedis(testRedis, hq)
				require.NoError(t, err)
			},
			verify: func(t *testing.T) {
				hqKey := bankKeyPrefix + "AGRIMCM1XXX"
				exists, err := testRedis.Exists(testCtx, hqKey).Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), exists, "headquarters should be deleted")

				isoKey := iso2IndexKey + ":MC"
				members, err := testRedis.SMembers(testCtx, isoKey).Result()
				require.NoError(t, err)
				assert.NotContains(t, members, hqKey, "ISO2 index should not contain HQ")
			},
			wantErr: false,
		},
		{
			name:        "nonexistent_bank",
			swiftPrefix: "NONEXIST",
			setup:       nil,
			verify: func(t *testing.T) {
				branchSetKey := "branch:NONEXIST"
				exists, err := testRedis.Exists(testCtx, branchSetKey).Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), exists, "branch set should not exist")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testRedis.FlushDB(testCtx).Err()
			require.NoError(t, err)

			if tt.setup != nil {
				tt.setup(t)
			}

			err = DeleteBanksBySwiftPrefix(testRedis, tt.swiftPrefix)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.verify != nil {
				tt.verify(t)
			}
		})
	}
}
