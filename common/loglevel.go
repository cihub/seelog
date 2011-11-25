// Package common contains constants and basic types for the sealog packages
package common

// Log level type
type LogLevel uint8

// Log levels
const (
	TraceLvl = iota
	DebugLvl
	InfoLvl
	WarnLvl
	ErrorLvl
	CriticalLvl
	Off
)

// Log level string representations (used in configuration files)
const (
	TraceStr    = "trace"
	DebugStr    = "debug"
	InfoStr     = "info"
	WarnStr     = "warn"
	ErrorStr    = "error"
	CriticalStr = "critical"
	OffStr      = "off"
)

var levelToStringRepresentations = map[LogLevel]string{
	TraceLvl:    TraceStr,
	DebugLvl:    DebugStr,
	InfoLvl:     InfoStr,
	WarnLvl:     WarnStr,
	ErrorLvl:    ErrorStr,
	CriticalLvl: CriticalStr,
	Off:         OffStr,
}

// LogLevelFromString parses a string and returns a corresponding log level, if sucessfull. 
func LogLevelFromString(levelStr string) (level LogLevel, found bool) {
	for lvl, lvlStr := range levelToStringRepresentations {
		if lvlStr == levelStr {
			return lvl, true
		}
	}

	return 0, false
}

// LogLevelToString returns sealog string representation for a specified level. Returns "" for invalid log levels.
func (level LogLevel) String() string {
	levelStr, ok := levelToStringRepresentations[level]
	if ok {
		return levelStr
	}

	return ""
}

