package app

import (
	"fmt"
	"github.com/z4dm42f5kp-ux/numb-form/internal/lookup"
	"io"
	"os"
	"runtime"
	"strings"
)

// Environment variables holding the optional free API keys.
const (
	KeyEnv       = "NUMBFORM_COMPANIES_HOUSE_KEY"
	SerperKeyEnv = "NUMBFORM_SERPER_KEY"
	TavilyKeyEnv = "NUMBFORM_TAVILY_KEY"
)

// LookupKeyFromEnv returns the optional Companies House key, trimmed.
func LookupKeyFromEnv() string { return strings.TrimSpace(os.Getenv(KeyEnv)) }

// LookupKeysFromEnv returns every optional free key from the environment.
func LookupKeysFromEnv() LookupConfig {
	return LookupConfig{
		CompaniesHouseKey: LookupKeyFromEnv(),
		SerperKey:         strings.TrimSpace(os.Getenv(SerperKeyEnv)),
		TavilyKey:         strings.TrimSpace(os.Getenv(TavilyKeyEnv)),
	}
}

// Keys prints the state of the optional free API keys and how to set them.
// Keys are read from the environment only; numb-form never stores them.
func Keys(out io.Writer) int {
	keys := []struct{ env, what, where string }{
		{SerperKeyEnv, "Serper: Google results for the searches, and Google Maps for addresses.", "https://serper.dev (2,500 free searches, no credit card)"},
		{TavilyKeyEnv, "Tavily: the last resort, a short AI answer with its sources (never taken as verified).", "https://tavily.com (1,000 free a month, no credit card)"},
		{KeyEnv, "UK Companies House: officers and registered office of UK companies.", "https://developer.company-information.service.gov.uk/ (free REST key)"},
	}
	fmt.Fprintln(out, "Optional free API keys (numb-form never stores keys and never uses paid services)")
	for _, k := range keys {
		state := "not set"
		if v := strings.TrimSpace(os.Getenv(k.env)); v != "" {
			state = "set (" + mask(v) + ")"
		}
		fmt.Fprintf(out, "\n  %s: %s\n      %s\n      Get it at %s\n", k.env, state, k.what, k.where)
	}
	switch runtime.GOOS {
	case "windows":
		fmt.Fprintln(out, "\nTo set a key for your Windows user (new PowerShell windows only):\n  setx NAME \"your-key\"")
	default:
		fmt.Fprintln(out, "\nTo set a key, add to your shell profile:\n  export NAME='your-key'")
	}
	return ExitOK
}

func mask(k string) string {
	if len(k) <= 4 {
		return "****"
	}
	return "****" + k[len(k)-4:]
}

// LookupConfig is re-exported for the command packages.
type LookupConfig = lookup.Config
