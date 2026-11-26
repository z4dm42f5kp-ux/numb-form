// Package i18n translates what numb-form says on the terminal into
// Spanish and German.
//
// In plain words: the questions and messages appear in the language the
// person chose. The text written into the forms stays in English.
//
// A message is looked up by its English text (the format string before
// its values are filled in). A message without a translation is shown in
// English, so a missing entry is never an error.
package i18n

import "strings"

// Languages lists the codes the programs accept, English first.
var Languages = []string{"en", "es", "de"}

// Names are the languages as their speakers write them.
var Names = map[string]string{"en": "English", "es": "Español", "de": "Deutsch"}

var catalogs = map[string]map[string]string{"es": es, "de": de}

// Valid reports whether code is a supported language.
func Valid(code string) bool {
	_, ok := Names[code]
	return ok
}

// Parse returns the language code for an answer such as "2", "es",
// "Español" or "deutsch", and false for anything else.
func Parse(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	for i, code := range Languages {
		if s == code || s == strings.ToLower(Names[code]) || s == string(rune('1'+i)) {
			return code, true
		}
	}
	switch s {
	case "espanol", "spanish", "german", "aleman", "alemán", "englisch", "inglés", "ingles":
		return map[string]string{"espanol": "es", "spanish": "es", "german": "de", "aleman": "de", "alemán": "de",
			"englisch": "en", "inglés": "en", "ingles": "en"}[s], true
	}
	return "", false
}

// T returns the translation of the English text s, or s.
func T(lang, s string) string {
	if t, ok := catalogs[lang][s]; ok {
		return t
	}
	return s
}
