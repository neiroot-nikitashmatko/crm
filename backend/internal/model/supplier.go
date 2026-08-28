package model

type Supplier struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	ContactPerson        string `json:"contactPerson"`
	Phone                string `json:"phone"`
	INN                  string `json:"inn"`
	KPP                  string `json:"kpp"`
	OGRN                 string `json:"ogrn"`
	LegalAddress         string `json:"legalAddress"`
	ActualAddress        string `json:"actualAddress"`
	BIK                  string `json:"bik"`
	SettlementAccount    string `json:"settlementAccount"`
	CorrespondentAccount string `json:"correspondentAccount"`
	CreatedAt            int64  `json:"createdAt"`
}

type UpsertSupplierInput struct {
	Name                 string `json:"name"`
	ContactPerson        string `json:"contactPerson"`
	Phone                string `json:"phone"`
	INN                  string `json:"inn"`
	KPP                  string `json:"kpp"`
	OGRN                 string `json:"ogrn"`
	LegalAddress         string `json:"legalAddress"`
	ActualAddress        string `json:"actualAddress"`
	BIK                  string `json:"bik"`
	SettlementAccount    string `json:"settlementAccount"`
	CorrespondentAccount string `json:"correspondentAccount"`
}
