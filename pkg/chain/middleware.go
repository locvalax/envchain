package chain

import (
	"log"
	"strings"
	"time"
)

// Source is a local alias for readability within this file.
type resolveFunc func(key string) (string, bool)

// WithLogging wraps a resolve function and logs each lookup.
func WithLogging(resolve resolveFunc, logger *log.Logger) resolveFunc {
	return func(key string) (string, bool) {
		start := time.Now()
		val, ok := resolve(key)
		if ok {
			logger.Printf("[envchain] resolved %q in %s", key, time.Since(start))
		} else {
			logger.Printf("[envchain] missing  %q in %s", key, time.Since(start))
		}
		return val, ok
	}
}

// WithRedaction wraps a resolve function and redacts sensitive values in logs.
// Keys matching any of the sensitive patterns will have their values masked.
func WithRedaction(resolve resolveFunc, sensitivePatterns []string) resolveFunc {
	return func(key string) (string, bool) {
		val, ok := resolve(key)
		if ok && isSensitive(key, sensitivePatterns) {
			return "***REDACTED***", true
		}
		return val, ok
	}
}

// WithDefault wraps a resolve function and returns a default value when a key
// is not found in any source.
func WithDefault(resolve resolveFunc, defaults map[string]string) resolveFunc {
	return func(key string) (string, bool) {
		if val, ok := resolve(key); ok {
			return val, true
		}
		if def, ok := defaults[key]; ok {
			return def, true
		}
		return "", false
	}
}

func isSensitive(key string, patterns []string) bool {
	upper := strings.ToUpper(key)
	for _, p := range patterns {
		if strings.Contains(upper, strings.ToUpper(p)) {
			return true
		}
	}
	return false
}
