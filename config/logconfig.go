package config

import (
	"os"
	"sealog/dispatchers"
	. "sealog/common"
)

// LogConfig stores logging configuration. Contains messages dispatcher, allowed log levels rules 
// (general constraints and exceptions), and messages formats (used by nodes of dispatcher tree)
type LogConfig struct {
	Constraints    LogLevelConstraints      // General log level rules (>min and <max, or set of allowed levels)
	Exceptions     []*LogLevelException     // Exceptions to general rules for specific files or funcs
	RootDispatcher dispatchers.DispatcherInterface // Root of output tree
	cache map[string]map[string]map[string]bool
}

func NewConfig(constraints LogLevelConstraints, exceptions []*LogLevelException, rootDispatcher dispatchers.DispatcherInterface) (*LogConfig, os.Error) {
	if constraints == nil {
		return nil, os.NewError("Constraints can not be nil")
	}
	if rootDispatcher == nil {
		return nil, os.NewError("RootDispatcher can not be nil")
	}
	
	config := new(LogConfig)
	config.Constraints = constraints
	config.Exceptions = exceptions
	config.RootDispatcher = rootDispatcher
	config.cache =  make(map[string]map[string]map[string]bool)
	
	return config, nil
}

func (config *LogConfig) IsAllowed(level LogLevel, context *LogContext) bool {
	funcMap, ok := config.cache[context.FullPath()]
	if !ok {
		funcMap = make(map[string]map[string]bool, 0)
		config.cache[context.FullPath()] = funcMap
	}
	
	levelMap, ok := funcMap[context.Func()]
	if !ok {
		levelMap = make(map[string]bool, 0)
		funcMap[context.Func()] = levelMap
	}
	
	isAllowValue, ok := levelMap[level.String()]
	if !ok {
		isAllowValue = config.isAllowed(level, context)
		levelMap[level.String()] = isAllowValue
	}
	
	return isAllowValue
}

// IsAllowed returns true if logging with specified log level is allowed in current context.
// If any of exception patterns match current context, then exception constraints are applied. Otherwise,
// the general constraints are used.
func (config *LogConfig) isAllowed(level LogLevel, context *LogContext) bool {
	allowed := config.Constraints.IsAllowed(level) // General rule

	// Exceptions:

	for _, exception := range config.Exceptions {
		if exception.MatchesContext(context) {
			return exception.IsAllowed(level)
		}
	}

	return allowed
}
