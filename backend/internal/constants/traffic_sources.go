package constants

// ManualTrafficSources — источники, доступные при ручном создании лида.
var ManualTrafficSources = []string{
	"Знал о производстве",
	"Авито (AutoFactory)",
	"Визитка(авточехлы)",
	"Авито (Автоатрибут)",
	"Авито (Автоателье)",
	"Яндекс карты",
	"Instagram",
	"Вконтакте",
	"2gis",
	"Оптовый клиент",
	"Увидел вывеску",
	"Порекомендовали нас",
}

const (
	AvitoChatTrafficSource = "Авито Чат"
	BeelineFallbackSource  = "Билайн"
)

func IsAllowedTrafficSource(value string) bool {
	for _, source := range ManualTrafficSources {
		if source == value {
			return true
		}
	}
	switch value {
	case AvitoChatTrafficSource, BeelineFallbackSource:
		return true
	default:
		// Источники, привязанные к номерам Билайн, тоже из ManualTrafficSources.
		return false
	}
}
