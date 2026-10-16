package app

import (
	"fmt"
	"github.com/z4dm42f5kp-ux/numb-form/internal/lookup"
	"io"
	"os"
	"runtime"
	"strings"
)

// KeyEnv is the environment variable holding the free Companies House key.
const KeyEnv = "NUMBFORM_COMPANIES_HOUSE_KEY"

// LookupKeyFromEnv returns the optional free API key, trimmed.
func LookupKeyFromEnv() string { return strings.TrimSpace(os.Getenv(KeyEnv)) }

// Keys prints the state of the optional free API keys and how to set them.
// Keys are read from the environment only; numb-form never stores them.
func Keys(out io.Writer) int {
	k := LookupKeyFromEnv()
	state := "not set (Companies House is skipped; every other source needs no key)"
	if k != "" {
		state = "set (" + mask(k) + ")"
	}
	fmt.Fprintf(out, `Optional free API keys

  %s
      UK Companies House, the official UK register (officers and
      registered office). Free: register at
      https://developer.company-information.service.gov.uk/ and create
      a "REST" API key. Status: %s

numb-form never stores keys and never uses paid services.
`, KeyEnv, state)
	switch runtime.GOOS {
	case "windows":
		fmt.Fprintf(out, "\nTo set it for your Windows user (new terminals only):\n  setx %s \"your-key\"\n", KeyEnv)
	default:
		fmt.Fprintf(out, "\nTo set it for this shell:\n  export %s='your-key'\n", KeyEnv)
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
