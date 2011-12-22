// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"errors"
	"regexp"
	"strings"
	"fmt"
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

	constraints LogLevelConstraints
}

// NewLogLevelException creates a new exception. 
func NewLogLevelException(funcPattern string, filePattern string, constraints LogLevelConstraints) (*LogLevelException, error) {
	if constraints == nil {
		return nil, errors.New("Constraints can not be nil")
	}

	exception := new(LogLevelException)

	err := exception.initFuncPatternParts(funcPattern)
	if err != nil {
		return nil, err
	}
	exception.funcPattern = strings.Join(exception.funcPatternParts, "")

	err = exception.initFilePatternParts(filePattern)
	if err != nil {
		return nil, err
	}
	exception.filePattern = strings.Join(exception.filePatternParts, "")

	exception.constraints = constraints

	return exception, nil
}

// MatchesContext returns true if context matches the patterns of this LogLevelException
func (logLevelEx *LogLevelException) MatchesContext(context *LogContext) bool {
	return logLevelEx.match(context.Func(), context.FullPath())
}

// IsAllowed returns true if log level is allowed according to the constraints of this LogLevelException
func (logLevelEx *LogLevelException) IsAllowed(level LogLevel) bool {
	return logLevelEx.constraints.IsAllowed(level)
}

// FuncPattern returns the function pattern of a exception
func (logLevelEx *LogLevelException) FuncPattern() string {
	return logLevelEx.funcPattern
}

// FuncPattern returns the file pattern of a exception
func (logLevelEx *LogLevelException) FilePattern() string {
	return logLevelEx.filePattern
}

// initFuncPatternParts checks whether the func filter has a correct format and splits funcPattern on parts
func (logLevelEx *LogLevelException) initFuncPatternParts(funcPattern string) (err error) {

	if funcFormatValidator.FindString(funcPattern) != funcPattern {
		return errors.New("Func path \"" + funcPattern + "\" contains incorrect symbols. Only a-z A-Z 0-9 _ * . allowed)")
	}

	logLevelEx.funcPatternParts = splitPattern(funcPattern)
	return nil
}

// Checks whether the file filter has a correct format and splits file patterns using splitPattern.
func (logLevelEx *LogLevelException) initFilePatternParts(filePattern string) (err error) {

	if fileFormatValidator.FindString(filePattern) != filePattern {
		return errors.New("File path \"" + filePattern + "\" contains incorrect symbols. Only a-z A-Z 0-9 \\ / _ * . allowed)")
	}

	logLevelEx.filePatternParts = splitPattern(filePattern)
	return err
}

func (logLevelEx *LogLevelException) match(funcPath string, filePath string) bool {
	if !stringMatchesPattern(logLevelEx.funcPatternParts, funcPath) {
		return false
	}
	return stringMatchesPattern(logLevelEx.filePatternParts, filePath)
}

func (logLevelEx *LogLevelException) String() string {
	str := fmt.Sprintf("Func: %s File: %s ", logLevelEx.funcPattern, logLevelEx.filePattern)

	if logLevelEx.constraints != nil {
		str += fmt.Sprintf("Constr: %s", logLevelEx.constraints)
	} else {
		str += "nil"
	}

	return str
}

// splitPattern splits pattern into strings and asterisks. Example: "ab*cde**f" -> ["ab", "*", "cde", "*", "f"]
func splitPattern(pattern string) []string {
	patternParts := make([]string, 0)
	var lastChar rune
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
