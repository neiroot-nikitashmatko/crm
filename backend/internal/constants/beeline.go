package constants

const BeelineNumberDomain = "@rnd.so.ims.mnc099.mcc250.3gppnetwork.org"

// BeelineTrafficSourceByPhone maps multi-channel phone digits to lead traffic source.
var BeelineTrafficSourceByPhone = map[string]string{
	"9613001616": "Знал о производстве",
	"9613015050": "Знал о производстве",
	"9662066959": "Визитка(авточехлы)",
	"9613195219": "Авито (Автоатрибут)",
	"9064545834": "Авито (AutoFactory)",
	"9064545866": "Авито (Автоателье)",
	"9034306767": "Яндекс карты",
	"9034363336": "Instagram",
	"9613011458": "Вконтакте",
	"9613011460": "2gis",
}

func BeelineTrafficSourceForPhoneDigits(digits string) string {
	if source, ok := BeelineTrafficSourceByPhone[digits]; ok {
		return source
	}
	return ""
}
