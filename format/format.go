// Copyright (c) 2012 - Cloud Instruments Co. Ltd.
// 
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met: 
// 
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer. 
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution. 
// 
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package format

import (
	"errors"
	"fmt"
	. "github.com/cihub/seelog/common"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	VerbSymbol         = '%'
	VerbSymbolString   = "%"
	VerbParameterStart = '('
	VerbParameterEnd   = ')'
	DateDefaultFormat  = "2006-01-02"
	TimeFormat         = "15:04:05"
)

var DefaultFormatter *Formatter

func init() {
	var err error
	DefaultFormatter, err = NewFormatter("%Ns [%Level] %Msg%n")
	if err != nil {
		fmt.Println("Error during DefaultFormatter creation: " + err.Error())
	}
}

type verbFunc func(message string, level LogLevel, context *LogContext) interface{}
type verbFuncCreator func(param string) verbFunc

var verbFuncs = map[string]verbFunc{
	"Level":    verbLevel,
	"Lev":      verbLev,
	"LEVEL":    verbLEVEL,
	"LEV":      verbLEV,
	"l":        verbl,
	"Msg":      verbMsg,
	"FullPath": verbFullPath,
	"File":     verbFile,
	"RelFile":  verbRelFile,
	"Func":     verbFunction,
	"Time":     verbTime,
	"Ns":       verbNs,
	"n":        verbn,
	"t":        verbt,
}

var verbFuncsParametrized = map[string]verbFuncCreator{
	"Date": createDateTimeVerbFunc,
}

// Formatter is used to write messages in a specific format, inserting such additional data
// as log level, date/time, etc.
type Formatter struct {
	fmtStringOriginal string
	fmtString         string
	verbFuncs         []verbFunc
}

// NewFormatter creates a new formatter using a format string
func NewFormatter(formatString string) (*Formatter, error) {
	newFormatter := new(Formatter)
	newFormatter.fmtStringOriginal = formatString

	err := newFormatter.buildVerbFuncs()
	if err != nil {
		return nil, err
	}

	return newFormatter, nil
}

func (formatter *Formatter) buildVerbFuncs() error {
	formatter.verbFuncs = make([]verbFunc, 0)
	var fmtString string
	for i := 0; i < len(formatter.fmtStringOriginal); i++ {
		char := formatter.fmtStringOriginal[i]
		if char != VerbSymbol {
			fmtString += string(char)
			continue
		}

		isEndOfStr := i == len(formatter.fmtStringOriginal)-1
		if isEndOfStr {
			return errors.New(fmt.Sprintf("Format error: %v - last symbol", VerbSymbolString))
		}

		isDoubledVerbSymbol := formatter.fmtStringOriginal[i+1] == VerbSymbol
		if isDoubledVerbSymbol {
			fmtString += VerbSymbolString
			i++
			continue
		}

		function, nextI, err := formatter.extractVerbFunc(i + 1)
		if err != nil {
			return err
		}

		fmtString += "%v"
		i = nextI
		formatter.verbFuncs = append(formatter.verbFuncs, function)
	}

	formatter.fmtString = fmtString
	return nil
}

func (formatter *Formatter) extractVerbFunc(index int) (verbFunc, int, error) {
	letterSequence := formatter.extractLetterSequence(index)
	if len(letterSequence) == 0 {
		return nil, 0, errors.New(fmt.Sprintf("Format error: lack of verb after %v. At %v", VerbSymbolString, index))
	}

	function, verbLength, ok := formatter.findVerbFunc(letterSequence)
	if ok {
		return function, index + verbLength - 1, nil
	}

	function, verbLength, ok = formatter.findVerbFuncParametrized(letterSequence, index)
	if ok {
		return function, index + verbLength - 1, nil
	}

	return nil, 0, errors.New("Format error: unrecognized verb at " + strconv.Itoa(index))
}

func (formatter *Formatter) extractLetterSequence(index int) string {
	letters := ""
	
	bytesToParse := []byte(formatter.fmtStringOriginal[index:])
	runeCount := utf8.RuneCount(bytesToParse)
	for i := 0; i < runeCount; i++ {
		rune, runeSize := utf8.DecodeRune(bytesToParse)
		bytesToParse = bytesToParse[runeSize:]

		if unicode.IsLetter(rune) {
			letters += string(rune)
		} else {
			break
		}
	}
	return letters
}

func (formatter *Formatter) findVerbFunc(letters string) (verbFunc, int, bool) {
	currentVerb := letters
	for i := 0; i < len(letters); i++ {
		function, ok := verbFuncs[currentVerb]
		if ok {
			return function, len(currentVerb), ok
		}
		currentVerb = currentVerb[:len(currentVerb)-1]
	}

	return nil, 0, false
}

func (formatter *Formatter) findVerbFuncParametrized(letters string, lettersStartIndex int) (verbFunc, int, bool) {
	currentVerb := letters
	for i := 0; i < len(letters); i++ {
		functionCreator, ok := verbFuncsParametrized[currentVerb]
		if ok {
			paramter := ""
			parameterLen := 0
			isVerbEqualsLetters := i == 0 // if not, then letter goes after verb, and verb is parameterless
			if isVerbEqualsLetters {
				userParamter := ""
				userParamter, parameterLen, ok = formatter.findparameter(lettersStartIndex + len(currentVerb))
				if ok {
					paramter = userParamter
				}
			}

			return functionCreator(paramter), len(currentVerb) + parameterLen, true
		}

		currentVerb = currentVerb[:len(currentVerb)-1]
	}

	return nil, 0, false
}

func (formatter *Formatter) findparameter(startIndex int) (string, int, bool) {
	if len(formatter.fmtStringOriginal) == startIndex || formatter.fmtStringOriginal[startIndex] != VerbParameterStart {
		return "", 0, false
	}

	endIndex := strings.Index(formatter.fmtStringOriginal[startIndex:], string(VerbParameterEnd)) + startIndex
	if endIndex == -1 {
		return "", 0, false
	}

	length := endIndex - startIndex + 1

	return formatter.fmtStringOriginal[startIndex+1 : endIndex], length, true
}

// Format processes a message with special verbs, log level, and context. Returns formatted string
// with all verb identifiers changed to appropriate values.
func (formatter *Formatter) Format(message string, level LogLevel, context *LogContext) string {
	if len(formatter.verbFuncs) == 0 {
		return formatter.fmtString
	}

	params := make([]interface{}, len(formatter.verbFuncs))
	for i, function := range formatter.verbFuncs {
		params[i] = function(message, level, context)
	}

	return fmt.Sprintf(formatter.fmtString, params...)
}

func (formatter *Formatter) String() string {
	return formatter.fmtStringOriginal
}

//=====================================================

const (
	wrongLogLevel = "WRONG_LOGLEVEL"
)

var levelToString = map[LogLevel]string{
	TraceLvl:    "Trace",
	DebugLvl:    "Debug",
	InfoLvl:     "Info",
	WarnLvl:     "Warn",
	ErrorLvl:    "Error",
	CriticalLvl: "Critical",
	Off:         "Off",
}

var levelToShortString = map[LogLevel]string{
	TraceLvl:    "Trc",
	DebugLvl:    "Dbg",
	InfoLvl:     "Inf",
	WarnLvl:     "Wrn",
	ErrorLvl:    "Err",
	CriticalLvl: "Crt",
	Off:         "Off",
}

var levelToShortestString = map[LogLevel]string{
	TraceLvl:    "t",
	DebugLvl:    "d",
	InfoLvl:     "i",
	WarnLvl:     "w",
	ErrorLvl:    "e",
	CriticalLvl: "c",
	Off:         "o",
}

func verbLevel(message string, level LogLevel, context *LogContext) interface{} {
	levelStr, ok := levelToString[level]
	if !ok {
		return wrongLogLevel
	}
	return levelStr
}

func verbLev(message string, level LogLevel, context *LogContext) interface{} {
	levelStr, ok := levelToShortString[level]
	if !ok {
		return wrongLogLevel
	}
	return levelStr
}

func verbLEVEL(message string, level LogLevel, context *LogContext) interface{} {
	return strings.ToTitle(verbLevel(message, level, context).(string))
}

func verbLEV(message string, level LogLevel, context *LogContext) interface{} {
	return strings.ToTitle(verbLev(message, level, context).(string))
}

func verbl(message string, level LogLevel, context *LogContext) interface{} {
	levelStr, ok := levelToShortestString[level]
	if !ok {
		return wrongLogLevel
	}
	return levelStr
}

func verbMsg(message string, level LogLevel, context *LogContext) interface{} {
	return message
}

func verbFullPath(message string, level LogLevel, context *LogContext) interface{} {
	return context.FullPath()
}

func verbFile(message string, level LogLevel, context *LogContext) interface{} {
	return context.FileName()
}

func verbRelFile(message string, level LogLevel, context *LogContext) interface{} {
	return context.ShortPath()
}

func verbFunction(message string, level LogLevel, context *LogContext) interface{} {
	return context.Func()
}

func verbTime(message string, level LogLevel, context *LogContext) interface{} {
	return context.CallTime().Format(TimeFormat)
}

func verbNs(message string, level LogLevel, context *LogContext) interface{} {
	return context.CallTime().UnixNano()
}

func verbn(message string, level LogLevel, context *LogContext) interface{} {
	return "\n"
}

func verbt(message string, level LogLevel, context *LogContext) interface{} {
	return "\t"
}

func createDateTimeVerbFunc(dateTimeFormat string) verbFunc {
	format := dateTimeFormat
	if format == "" {
		format = DateDefaultFormat
	}
	return func(message string, level LogLevel, context *LogContext) interface{} {
		return time.Now().Format(format)
	}
}
