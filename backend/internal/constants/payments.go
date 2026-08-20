package constants

var PaymentPayerLabels = map[string]string{
	"ip-panov-nikolay":  "ИП Панов Николай Константинович",
	"ip-shmatko-nikita": "ИП Шматко Никита Сергеевич",
	"ip-panov-dmitry":   "ИП Панов Дмитрий Константинович",
}

func PaymentPayerLabel(payerID string) string {
	if label, ok := PaymentPayerLabels[payerID]; ok {
		return label
	}
	return payerID
}
