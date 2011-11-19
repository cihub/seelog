package sealog

import (
	"sealog/common"
	"os"
	"regexp"
	"strings"
)

// Used in rules creation to validate input file and func filters
var (
	fileFormatValidator = regexp.MustCompile(`[a-zA-Z0-9\\/ _\*\.]*`)
	funcFormatValidator = regexp.MustCompile(`[a-zA-Z0-9_\*\.]*`)
)

// LogLevelException represents an exceptional case used when you need some specific files or funcs to
// override general constraints and to use their own.
type LogLevelException struct {
	funcPatternParts []string
	filePatternParts []string

	funcPattern string
	filePattern string

	contraints LogLevelConstraints
}

// NewLogLevelException creates a new exception. 
func NewLogLevelException(funcPattern string, filePattern string, contraints LogLevelConstraints) (*LogLevelException, os.Error) {
	if contraints == nil {
		return nil, os.NewError("Constraints can not be nil")
	}

	exception := new(LogLevelException)
	exception.funcPattern = funcPattern
	exception.filePattern = filePattern

	err := exception.initFuncPatternParts(funcPattern)
	if err != nil {
		return nil, err
	}

	err = exception.initFilePatternParts(filePattern)
	if err != nil {
		return nil, err
	}

	return exception, nil
}

// MatchesContext returns true if context matches the patterns of this LogLevelException
func (this *LogLevelException) MatchesContext(context *common.LogContext) bool {
	return true
}

// IsAllowed returns true if log level is allowed according to the constraints of this LogLevelException
func (this *LogLevelException) IsAllowed(level common.LogLevel) bool {
	return this.contraints.IsAllowed(level)
}

// initFuncPatternParts checks whether the func filter has a correct format and splits funcPattern on parts
func (this *LogLevelException) initFuncPatternParts(funcPattern string) (error os.Error) {

	if funcFormatValidator.FindString(funcPattern) != funcPattern {
		return os.NewError("Func path \"" + funcPattern + "\" contains incorrect symbols. Only a-z A-Z 0-9 _ * . allowed)")
	}

	this.funcPatternParts = splitPattern(funcPattern)
	return nil
}

// Checks whether the file filter has a correct format and splits file patterns using splitPattern.
func (this *LogLevelException) initFilePatternParts(filePattern string) (error os.Error) {

	if fileFormatValidator.FindString(filePattern) != filePattern {
		return os.NewError("File path \"" + filePattern + "\" contains incorrect symbols. Only a-z A-Z 0-9 \\ / _ * . allowed)")
	}

	this.filePatternParts = splitPattern(filePattern)
	return error
}

func (this *LogLevelException) match(funcPath string, filePath string) bool {
	if !stringMatchesPattern(this.funcPatternParts, funcPath) {
		return false
	}
	return stringMatchesPattern(this.filePatternParts, filePath)
}

// splitPattern splits pattern into strings and asterisks. Example: "ab*cde**f" -> ["ab", "*", "cde", "*", "f"]
func splitPattern(pattern string) []string {
	patternParts := make([]string, 0)
	var lastChar int
	for _, char := range pattern {
		if char == '*' {
			if lastChar != '*' {
				patternParts = append(patternParts, "*")
			}
		} else {
			if len(patternParts) != 0 && lastChar != '*' {
				patternParts[len(patternParts)-1] += string(char)
			} else {
				patternParts = append(patternParts, string(char))
			}
		}
		lastChar = char
	}

	return patternParts
}

// stringMatchesPattern check whether testString matches pattern with asterisks.
// Standard regexp functionality is not used here because of performance issues.
func stringMatchesPattern(patternparts []string, testString string) bool {
	if len(patternparts) == 0 {
		return len(testString) == 0
	}

	part := patternparts[0]
	if part != "*" {
		index := strings.Index(testString, part)
		if index == 0 {
			return stringMatchesPattern(patternparts[1:], testString[len(part):])
		}
	} else {
		if len(patternparts) == 1 {
			return true
		}

		newTestString := testString
		part = patternparts[1]
		for {
			index := strings.Index(newTestString, part)
			if index == -1 {
				break
			}

			newTestString = newTestString[index+len(part):]
			result := stringMatchesPattern(patternparts[2:], newTestString)
			if result {
				return true
			}
		}
	}
	return false
}
