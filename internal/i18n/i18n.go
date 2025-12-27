package i18n

import (
	"encoding/json"
	"os"
	"strings"
	"sync"

	_ "embed"

	"github.com/game_tool_box/internal/config"
)

type Lang string

const (
	LangZH Lang = "zh"
	LangEN Lang = "en"
)

//go:embed locales/zh.json
var zhJSON []byte

//go:embed locales/en.json
var enJSON []byte

var (
	zhCN map[string]string
	enUS map[string]string

	mu      sync.RWMutex
	current Lang
)

func init() {
	// Keep init panic-free: if parsing fails, fall back to empty maps.
	zhCN = map[string]string{}
	enUS = map[string]string{}

	_ = json.Unmarshal(zhJSON, &zhCN)
	_ = json.Unmarshal(enJSON, &enUS)

	current = Detect()
	if c, err := config.Load(); err == nil {
		if strings.TrimSpace(c.Lang) != "" {
			current = Lang(strings.ToLower(strings.TrimSpace(c.Lang)))
		}
	}
}

// Supported returns all supported UI languages.
func Supported() []Lang {
	// Keep ordering stable for UI.
	return []Lang{LangZH, LangEN}
}

// Current returns current UI language.
func Current() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// SetCurrent switches UI language at runtime.
func SetCurrent(lang Lang) {
	mu.Lock()
	current = lang
	mu.Unlock()
}

// SetCurrentPersisted switches UI language at runtime and persists it.
func SetCurrentPersisted(lang Lang) {
	SetCurrent(lang)
	// Best-effort persistence (do not block UI if it fails).
	c, _ := config.Load()
	c.Lang = string(lang)
	_ = config.Save(c)
}

// LangName returns the display name of a language (localized to `inLang`).
func LangName(inLang Lang, lang Lang) string {
	return T(inLang, "lang."+string(lang))
}

// Detect returns the best-effort UI language.
//
// Order:
//  1. GAMETOOLBOX_LANG env ("zh" / "en")
//  2. OS locale env (LANG / LC_ALL etc.) for prefixes "zh" or "en"
//  3. fallback: en
func Detect() Lang {
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("GAMETOOLBOX_LANG"))); v != "" {
		if strings.HasPrefix(v, "zh") {
			return LangZH
		}
		if strings.HasPrefix(v, "en") {
			return LangEN
		}
	}

	// Common locale envs
	for _, k := range []string{"LANG", "LC_ALL", "LC_MESSAGES"} {
		v := strings.ToLower(strings.TrimSpace(os.Getenv(k)))
		if strings.HasPrefix(v, "zh") {
			return LangZH
		}
		if strings.HasPrefix(v, "en") {
			return LangEN
		}
	}

	return LangEN
}

// T returns a localized string for the given key.
func T(lang Lang, key string) string {
	if lang == LangZH {
		if v, ok := zhCN[key]; ok {
			return v
		}
	}
	if v, ok := enUS[key]; ok {
		return v
	}
	// fallback to key for dev visibility
	return key
}
