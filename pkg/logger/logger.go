package logger

import (
	"fmt"
	"log"
)

var (
	DebugEnabled   bool
	VerboseEnabled bool
)

func Debug(format string, args ...any) {
	if !DebugEnabled {
		return
	}
	log.Printf("[DEBUG] "+format, args...)
}

func Verbose(format string, args ...any) {
	if !DebugEnabled || !VerboseEnabled {
		return
	}
	log.Printf("[VERBOSE] "+format, args...)
}

// MaskToken shows the first 10 chars then *** to avoid leaking credentials in logs.
func MaskToken(token string) string {
	if len(token) == 0 {
		return "<empty>"
	}
	if len(token) <= 10 {
		return "***"
	}
	return fmt.Sprintf("%s***", token[:10])
}
