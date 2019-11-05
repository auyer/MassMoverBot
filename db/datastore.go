package db

// DataStorage interface defines the methods required by the bot to store data
type DataStorage interface {
	Close() error
	GetStatistics() (map[string]int, error)
	SetStatistics(map[string]int) error
	WasWelcomeMessageSent(id string) (bool, error)
	SetWelcomeMessageSent(id string, value bool) error
	GetGuildLang(id string) (string, error)
	SetGuildLang(id, value string) error
	DeleteGuildLang(id string) error
}
