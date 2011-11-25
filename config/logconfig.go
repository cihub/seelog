package config

import (
	"sealog/dispatchers"
	. "sealog/common"
)

// LogConfig stores logging configuration. Contains messages dispatcher, allowed log levels rules 
// (general constraints and exceptions), and messages formats (used by nodes of dispatcher tree)
type LogConfig struct {
	Constraints    LogLevelConstraints      // General log level rules (>min and <max, or set of allowed levels)
	Exceptions     []*LogLevelException     // Exceptions to general rules for specific files or funcs
	RootDispatcher dispatchers.DispatcherInterface // Root of output tree
}

// IsAllowed returns true if logging with specified log level is allowed in current context.
// If any of exception patterns match current context, then exception constraints are applied. Otherwise,
// the general constraints are used.
func (config *LogConfig) IsAllowed(level LogLevel, context *LogContext) bool {
	allowed := config.Constraints.IsAllowed(level) // General rule

	// Exceptions:

	for _, exception := range config.Exceptions {
		if exception.MatchesContext(context) {
			return exception.IsAllowed(level)
		}
	}

	return allowed
}
