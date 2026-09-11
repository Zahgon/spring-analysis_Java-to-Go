// Package local holds the localisation demonstration.
package local

import (
	"fmt"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
)

// BundleBaseName is the bundle the demonstration reads. The original notes
// that the file must sit on the classpath; here that is the resources root.
const BundleBaseName = "resource/resource"

// GreetingKey is the key the demonstration looks up.
const GreetingKey = "greeting.common"

// ResourceBundleDemo reads the greeting in en_US and then in zh_CN and prints
// each, prefixed with the locale's country.
func ResourceBundleDemo(classPath string) error {
	for _, l := range []struct {
		label  string
		locale javart.Locale
	}{
		{"US", javart.LocaleUS},
		{"CN", javart.LocaleChina},
	} {
		bundle, err := javart.GetBundle(classPath, BundleBaseName, l.locale)
		if err != nil {
			return err
		}
		greeting, err := bundle.GetString(GreetingKey)
		if err != nil {
			return err
		}
		fmt.Println(l.label + ": " + greeting)
	}
	return nil
}
