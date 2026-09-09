package profile

import (
	_ "embed"
	"strings"
	"sync"
)

//go:embed profile.enc
var embeddedProfileCiphertext string

var (
	cachedPayload *Payload
	loadOnce      sync.Once
	loadErr       error
)

// ensureLoaded decrypts and caches the developer profile payload.
func ensureLoaded() (*Payload, error) {
	loadOnce.Do(func() {
		raw := strings.TrimSpace(embeddedProfileCiphertext)
		cachedPayload, loadErr = DecryptPayload(raw)
	})
	return cachedPayload, loadErr
}

// GetProfile returns the localized profile and common links for the given language.
// Defaults to English if the requested language is not available.
func GetProfile(lang string) (LangProfile, Links) {
	payload, err := ensureLoaded()
	if err != nil || payload == nil {
		return fallbackProfile(lang), Links{
			Web:      "https://retry.day",
			Git:      "https://github.com/sorrynthx/leitstand.git",
			LinkedIn: "https://www.linkedin.com/in/kyunggon-kim",
		}
	}

	norm := strings.ToLower(strings.TrimSpace(lang))
	if norm != "ko" && norm != "de" {
		norm = "en"
	}

	prof, ok := payload.Profiles[norm]
	if !ok {
		prof = payload.Profiles["en"]
	}

	return prof, payload.Links
}

func fallbackProfile(lang string) LangProfile {
	switch strings.ToLower(lang) {
	case "ko":
		return LangProfile{
			Name:        "김경곤 (Kyunggon Kim)",
			StoryTitle:  "삶과 사람에 대한 이야기",
			CareerTitle: "🤝 문제 해결 & 팀 동료",
			TechTitle:   "🛠️ Engineering Philosophy",
		}
	case "de":
		return LangProfile{
			Name:        "Kyunggon Kim",
			StoryTitle:  "Über das Leben & Menschen",
			CareerTitle: "🇩🇪 Karriere & Vor-Ort-Team in Deutschland",
			TechTitle:   "🛠️ Technische Philosophie",
		}
	default:
		return LangProfile{
			Name:        "Kyunggon Kim",
			StoryTitle:  "Life & People",
			CareerTitle: "🤝 Problem Solver & Trusted Teammate",
			TechTitle:   "🛠️ Engineering Philosophy",
		}
	}
}
