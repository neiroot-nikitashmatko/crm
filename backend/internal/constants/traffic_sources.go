package constants

// ManualTrafficSources — источники, доступные при ручном создании лида.
var ManualTrafficSources = []string{
	"Знал о производстве",
	"Авито (Автоатрибут)",
	"Авито (AutoFactory)",
	"Авито (Автоателье)",
	"По рекомендации",
	"Увидел вывеску",
	"Оптовый клиент",
	"Яндекс карты",
	"Instagram",
	"Вконтакте",
	"2gis",
	"Визитка (установка чехлов)",
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
