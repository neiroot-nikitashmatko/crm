package model

type TrafficSourceMetric struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

type LeadToDealConversion struct {
	LeadsCount     int     `json:"leadsCount"`
	ConvertedCount int     `json:"convertedCount"`
	Percent        float64 `json:"percent"`
}

type FailedLeadShare struct {
	LeadsCount  int     `json:"leadsCount"`
	FailedCount int     `json:"failedCount"`
	Percent     float64 `json:"percent"`
}

type FailedDealShare struct {
	DealsCount  int     `json:"dealsCount"`
	FailedCount int     `json:"failedCount"`
	Percent     float64 `json:"percent"`
}

type NomenclatureCount struct {
	Nomenclature string `json:"nomenclature"`
	Count        int    `json:"count"`
}

type ProductionCategoryMetric struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type EmployeeShareMetric struct {
	Employee string `json:"employee"`
	Count    int    `json:"count"`
}
