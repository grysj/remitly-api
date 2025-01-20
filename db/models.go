package db

type RedisRow struct {
	ISO2       string `json:"countryISO2" redis:"countryISO2"`
	Name       string `json:"bankName" redis:"bankName"`
	Type       string `json:"type,omitempty" redis:"type"`
	Address    string `json:"address,omitempty" redis:"address"`
	Town       string `json:"town,omitempty" redis:"town"`
	Country    string `json:"countryName,omitempty" redis:"countryName"`
	Timezone   string `json:"timezone,omitempty" redis:"timezone"`
	Headquater bool   `json:"isHeadquater" redis:"isHeadquater"`
}

type Bank struct {
	Swift      string `json:"swiftCode" redis:"swiftCode"`
	ISO2       string `json:"countryISO2" redis:"countryISO2"`
	Name       string `json:"bankName" redis:"bankName"`
	Type       string `json:"type,omitempty" redis:"type"`
	Address    string `json:"address,omitempty" redis:"address"`
	Town       string `json:"town,omitempty" redis:"town"`
	Country    string `json:"countryName,omitempty" redis:"countryName"`
	Timezone   string `json:"timezone,omitempty" redis:"timezone"`
	Headquater bool   `json:"isHeadquater" redis:"isHeadquater"`
}

type DeleteBankParams struct {
	Swift string `json:"swiftCode" redis:"swiftCode"`
}

type GetBankByIsoResult struct {
	Swift      string `json:"swiftCode" redis:"swiftCode"`
	ISO2       string `json:"countryISO2" redis:"countryISO2"`
	Name       string `json:"bankName" redis:"bankName"`
	Type       string `json:"type,omitempty" redis:"-"`
	Address    string `json:"address,omitempty" redis:"address"`
	Town       string `json:"town,omitempty" redis:"-"`
	Country    string `json:"countryName" redis:"-"`
	Timezone   string `json:"timezone,omitempty" redis:"-"`
	Headquater bool   `json:"isHeadquater" redis:"isHeadquater"`
}

type GetBranchesBySwiftResult struct {
	Swift      string `json:"swiftCode" redis:"swiftCode"`
	ISO2       string `json:"countryISO2" redis:"countryISO2"`
	Name       string `json:"bankName" redis:"bankName"`
	Type       string `json:"type,omitempty" redis:"-"`
	Address    string `json:"address,omitempty" redis:"address"`
	Town       string `json:"town,omitempty" redis:"-"`
	Country    string `json:"countryName,omitempty" redis:"-"`
	Timezone   string `json:"timezone,omitempty" redis:"-"`
	Headquater bool   `json:"isHeadquater" redis:"isHeadquater"`
}

type GetBankBySwiftResult struct {
	Swift      string `json:"swiftCode" redis:"swiftCode"`
	ISO2       string `json:"countryISO2" redis:"countryISO2"`
	Name       string `json:"bankName" redis:"bankName"`
	Type       string `json:"type,omitempty" redis:"-"`
	Address    string `json:"address,omitempty" redis:"address"`
	Town       string `json:"town,omitempty" redis:"-"`
	Country    string `json:"countryName,omitempty" redis:"countryName"`
	Timezone   string `json:"timezone,omitempty" redis:"-"`
	Headquater bool   `json:"isHeadquater" redis:"isHeadquater"`
}
