package javart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Locale mirrors java.util.Locale for the two locales the application uses.
type Locale struct {
	Language string
	Country  string
}

// The locale constants the application refers to by name.
var (
	LocaleUS    = Locale{Language: "en", Country: "US"}
	LocaleChina = Locale{Language: "zh", Country: "CN"}
)

// String renders the locale as Java does: "en_US".
func (l Locale) String() string {
	if l.Country == "" {
		return l.Language
	}
	return l.Language + "_" + l.Country
}

// ResourceBundle is a loaded .properties bundle.
type ResourceBundle struct {
	baseName string
	locale   Locale
	props    Properties
}

// MissingResourceError is returned when no candidate file exists for a bundle,
// mirroring java.util.MissingResourceException.
type MissingResourceError struct {
	BaseName string
	Locale   Locale
}

func (e *MissingResourceError) Error() string {
	return fmt.Sprintf("Can't find bundle for base name %s, locale %s", e.BaseName, e.Locale)
}

// GetBundle loads baseName for locale from root, applying Java's candidate
// order: baseName_lang_COUNTRY, then baseName_lang, then baseName. root plays
// the part of the classpath.
func GetBundle(root, baseName string, locale Locale) (*ResourceBundle, error) {
	for _, suffix := range candidateSuffixes(locale) {
		path := filepath.Join(root, filepath.FromSlash(baseName)+suffix+".properties")
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		props, err := LoadProperties(f)
		closeErr := f.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		return &ResourceBundle{baseName: baseName, locale: locale, props: props}, nil
	}
	return nil, &MissingResourceError{BaseName: baseName, Locale: locale}
}

func candidateSuffixes(l Locale) []string {
	var out []string
	if l.Language != "" && l.Country != "" {
		out = append(out, "_"+l.Language+"_"+l.Country)
	}
	if l.Language != "" {
		out = append(out, "_"+l.Language)
	}
	return append(out, "")
}

// GetString returns the value for key. A key with no entry is an error, as
// java.util.ResourceBundle throws MissingResourceException.
func (b *ResourceBundle) GetString(key string) (string, error) {
	v, ok := b.props[key]
	if !ok {
		return "", fmt.Errorf("Can't find resource for bundle %s, key %s", b.baseName, key)
	}
	return v, nil
}

// Keys returns the bundle's keys.
func (b *ResourceBundle) Keys() []string {
	keys := make([]string, 0, len(b.props))
	for k := range b.props {
		keys = append(keys, k)
	}
	return keys
}

// Locale returns the locale the bundle was requested for.
func (b *ResourceBundle) Locale() Locale { return b.locale }

// ParseLocale accepts the "lang_COUNTRY" form Java prints.
func ParseLocale(s string) Locale {
	parts := strings.SplitN(s, "_", 2)
	l := Locale{Language: parts[0]}
	if len(parts) == 2 {
		l.Country = parts[1]
	}
	return l
}
