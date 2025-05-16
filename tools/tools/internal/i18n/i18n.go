package i18n

import (
	"fmt"
	"os"
	"strings"
)

const (
	// EN is English.
	EN = "en"
	// PT is Portuguese.
	PT = "pt"
)

// Catalog looks up translated CLI strings.
type Catalog struct {
	locale string
}

// New returns a catalog for locale (en or pt).
func New(locale string) Catalog {
	return Catalog{locale: Parse(locale)}
}

// Locale returns the resolved language code.
func (c Catalog) Locale() string {
	if c.locale == "" {
		return EN
	}
	return c.locale
}

// T returns the translation for key, optionally formatted with args.
func (c Catalog) T(key string, args ...any) string {
	s := lookup(c.Locale(), key)
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

// Parse maps a tag like pt-BR or pt_BR.UTF-8 to en or pt.
func Parse(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return EN
	}
	raw, _, _ = strings.Cut(raw, ".")
	base, _, _ := strings.Cut(raw, "-")
	base, _, _ = strings.Cut(base, "_")
	if base == PT {
		return PT
	}
	return EN
}

// FromEnv resolves TATUSCAN_LANG from the process environment (typically .env).
func FromEnv() Catalog {
	if v := strings.TrimSpace(os.Getenv("TATUSCAN_LANG")); v != "" {
		return New(v)
	}
	return New(EN)
}

func lookup(locale, key string) string {
	if locale == PT {
		if s, ok := pt[key]; ok {
			return s
		}
	}
	if s, ok := en[key]; ok {
		return s
	}
	return key
}
