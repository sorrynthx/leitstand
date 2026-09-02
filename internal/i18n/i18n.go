package i18n

import (
	"fmt"
	"sync"
)

type Lang string

const (
	LangEN Lang = "en"
	LangKO Lang = "ko"
	LangDE Lang = "de"
)

type LangOption struct {
	Code  Lang
	Label string
}

var (
	mu          sync.RWMutex
	currentLang = LangEN

	SupportedLangs = []LangOption{
		{Code: LangEN, Label: "🇺🇸 English"},
		{Code: LangKO, Label: "🇰🇷 한국어 (Korean)"},
		{Code: LangDE, Label: "🇩🇪 Deutsch (German)"},
	}

	messages = map[Lang]map[string]string{
		LangEN: dictEN,
		LangKO: dictKO,
		LangDE: dictDE,
	}
)

func SetLang(l Lang) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := messages[l]; exists {
		currentLang = l
	}
}

func GetLang() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if dict, exists := messages[currentLang]; exists {
		if val, found := dict[key]; found {
			return val
		}
	}

	if dict, exists := messages[LangEN]; exists {
		if val, found := dict[key]; found {
			return val
		}
	}

	return key
}

func Tf(key string, args ...interface{}) string {
	return fmt.Sprintf(T(key), args...)
}
