package i18n

import (
	"fmt"
	"strings"
)

const (
	// EN is English.
	EN = "en"
	// PT is Portuguese.
	PT = "pt"
)

// Catalog looks up translated UI strings.
type Catalog struct {
	locale string
}

// New returns a catalog for locale (en or pt). Unknown values fall back to en.
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

// HTMLLang is the BCP 47 tag for the html lang attribute.
func (c Catalog) HTMLLang() string {
	if c.Locale() == PT {
		return "pt-BR"
	}
	return EN
}

// T returns the translation for key, optionally formatted with args.
func (c Catalog) T(key string, args ...any) string {
	s := lookup(c.Locale(), key)
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

// Parse maps a tag like pt-BR or en-US to en or pt.
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
