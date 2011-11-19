package sealog

import (
	"sealog/dispatchers"
	"sealog/common"
	//"sealog/format"
)

// Logging configuration. Contains messages dispatcher, allowed log levels selection rules 
// (general constraints and exceptions), and messages formats (used by nodes of dispatcher tree)
type LogConfig struct {
	Constraints    LogLevelConstraints             // General log level rules (>min and <max, or set of allowed levels)
	Exceptions     []*LogLevelException            // Exceptions to general rules for specific files or funcs
	RootDispatcher dispatchers.DispatcherInterface // Root of output tree
	//Formats []format.Format // Log formats	
}

/*// ConfigFromFile creates a config from file. File should contain valid sealog xml.
func ConfigFromFile(fileName string) (config *LogConfig, error os.Error) {
	...
}

// ConfigFromBytes creates a config from bytes stream. Bytes should contain valid sealog xml.
func ConfigFromBytes(bytes []byte) (config *LogConfig, error os.Error) {
	...
}

// ConfigForCompatibility creates a simple config for usage with non-Sealog systems. Configures system to write to output with minimal level = minLevel.
func ConfigForCompatibility(output io.Writer, minLevel LogLevel) (config *LogConfig, error os.Error) {
	...
}*/

// IsAllowed returns true if logging with specified log level is allowed in current context.
// If any of exception patterns match current context, then exception constraints are applied. Otherwise,
// the general constraints are used.
func (this *LogConfig) IsAllowed(level common.LogLevel, context *common.LogContext) bool {
	allowed := this.Constraints.IsAllowed(level) // General rule

	// Exceptions:

	for _, exception := range this.Exceptions {
		if exception.MatchesContext(context) {
			return exception.IsAllowed(level)
		}
	}

	return allowed
}
