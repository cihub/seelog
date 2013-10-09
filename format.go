// Copyright (c) 2012 - Cloud Instruments Co., Ltd.
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

package seelog

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// VerbSymbol is a special symbol used in config files to mark special format aliases.
const (
	VerbSymbol = '%'
)
const (
	verbSymbolString   = "%"
	verbParameterStart = '('
	verbParameterEnd   = ')'
)

// These are the time and date formats that are used when %Date or %Time format aliases are used.
const (
	DateDefaultFormat = "2006-01-02"
	TimeFormat        = "15:04:05"
)

var DefaultMsgFormat = "%Ns [%Level] %Msg%n"

var defaultformatter *formatter
var msgonlyformatter *formatter

func init() {
	var err error
	defaultformatter, err = newFormatter(DefaultMsgFormat)
	if err != nil {
		fmt.Println("Error during defaultformatter creation: " + err.Error())
	}
	msgonlyformatter, err = newFormatter("%Msg")
	if err != nil {
		fmt.Println("Error during msgonlyformatter creation: " + err.Error())
	}
}

type verbFunc func(message string, level LogLevel, context LogContextInterface) interface{}
type verbFuncCreator func(param string) verbFunc

var verbFuncs = map[string]verbFunc{
	"Level":     verbLevel,
	"Lev":       verbLev,
	"LEVEL":     verbLEVEL,
	"LEV":       verbLEV,
	"l":         verbl,
	"Msg":       verbMsg,
	"FullPath":  verbFullPath,
	"File":      verbFile,
	"RelFile":   verbRelFile,
	"Func":      verbFunction,
	"FuncShort": verbFunctionShort,
	"Line":      verbLine,
	"Time":      verbTime,
	"Ns":        verbNs,
	"n":         verbn,
	"t":         verbt,
}

var verbFuncsParametrized = map[string]verbFuncCreator{
	"Date": createDateTimeVerbFunc,
	"EscM": createANSIEscapeFunc,
}

// formatter is used to write messages in a specific format, inserting such additional data
// as log level, date/time, etc.
type formatter struct {
	fmtStringOriginal string
	fmtString         string
	verbFuncs         []verbFunc
}

// newFormatter creates a new formatter using a format string
func newFormatter(formatString string) (*formatter, error) {
	newformatter := new(formatter)
	newformatter.fmtStringOriginal = formatString

	err := newformatter.buildVerbFuncs()
	if err != nil {
		return nil, err
	}

	return newformatter, nil
}

func (formatter *formatter) buildVerbFuncs() error {
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
			return errors.New(fmt.Sprintf("Format error: %v - last symbol", verbSymbolString))
		}

		isDoubledVerbSymbol := formatter.fmtStringOriginal[i+1] == VerbSymbol
		if isDoubledVerbSymbol {
			fmtString += verbSymbolString
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

func (formatter *formatter) extractVerbFunc(index int) (verbFunc, int, error) {
	letterSequence := formatter.extractLetterSequence(index)
	if len(letterSequence) == 0 {
		return nil, 0, errors.New(fmt.Sprintf("Format error: lack of verb after %v. At %v", verbSymbolString, index))
	}

	function, verbLength, ok := formatter.findVerbFunc(letterSequence)
	if ok {
		return function, index + verbLength - 1, nil
	}

	function, verbLength, ok = formatter.findVerbFuncParametrized(letterSequence, index)
	if ok {
		return function, index + verbLength - 1, nil
	}

	return nil, 0, errors.New("Format error: unrecognized verb at " + strconv.Itoa(index) + ": " + letterSequence)
}

func (formatter *formatter) extractLetterSequence(index int) string {
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

func (formatter *formatter) findVerbFunc(letters string) (verbFunc, int, bool) {
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

func (formatter *formatter) findVerbFuncParametrized(letters string, lettersStartIndex int) (verbFunc, int, bool) {
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

func (formatter *formatter) findparameter(startIndex int) (string, int, bool) {
	if len(formatter.fmtStringOriginal) == startIndex || formatter.fmtStringOriginal[startIndex] != verbParameterStart {
		return "", 0, false
	}

	endIndex := strings.Index(formatter.fmtStringOriginal[startIndex:], string(verbParameterEnd)) + startIndex
	if endIndex == -1 {
		return "", 0, false
	}

	length := endIndex - startIndex + 1

	return formatter.fmtStringOriginal[startIndex+1 : endIndex], length, true
}

// Format processes a message with special verbs, log level, and context. Returns formatted string
// with all verb identifiers changed to appropriate values.
func (formatter *formatter) Format(message string, level LogLevel, context LogContextInterface) string {
	if len(formatter.verbFuncs) == 0 {
		return formatter.fmtString
	}

	params := make([]interface{}, len(formatter.verbFuncs))
	for i, function := range formatter.verbFuncs {
		params[i] = function(message, level, context)
	}

	return fmt.Sprintf(formatter.fmtString, params...)
}

func (formatter *formatter) String() string {
	return formatter.fmtStringOriginal
}

//=====================================================

const (
	wrongLogLevel   = "WRONG_LOGLEVEL"
	wrongEscapeCode = "WRONG_ESCAPE"
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

func verbLevel(message string, level LogLevel, context LogContextInterface) interface{} {
	levelStr, ok := levelToString[level]
	if !ok {
		return wrongLogLevel
	}
	return levelStr
}

func verbLev(message string, level LogLevel, context LogContextInterface) interface{} {
	levelStr, ok := levelToShortString[level]
	if !ok {
		return wrongLogLevel
	}
	return levelStr
}

func verbLEVEL(message string, level LogLevel, context LogContextInterface) interface{} {
	return strings.ToTitle(verbLevel(message, level, context).(string))
}

func verbLEV(message string, level LogLevel, context LogContextInterface) interface{} {
	return strings.ToTitle(verbLev(message, level, context).(string))
}

func verbl(message string, level LogLevel, context LogContextInterface) interface{} {
	levelStr, ok := levelToShortestString[level]
	if !ok {
		return wrongLogLevel
	}
	return levelStr
}

func verbMsg(message string, level LogLevel, context LogContextInterface) interface{} {
	return message
}

func verbFullPath(message string, level LogLevel, context LogContextInterface) interface{} {
	return context.FullPath()
}

func verbFile(message string, level LogLevel, context LogContextInterface) interface{} {
	return context.FileName()
}

func verbRelFile(message string, level LogLevel, context LogContextInterface) interface{} {
	return context.ShortPath()
}

func verbFunction(message string, level LogLevel, context LogContextInterface) interface{} {
	return context.Func()
}

func verbFunctionShort(message string, level LogLevel, context LogContextInterface) interface{} {
	f := context.Func()
	spl := strings.Split(f, ".")
	return spl[len(spl)-1]
}

func verbLine(message string, level LogLevel, context LogContextInterface) interface{} {
	return context.Line()
}

func verbTime(message string, level LogLevel, context LogContextInterface) interface{} {
	return context.CallTime().Format(TimeFormat)
}

func verbNs(message string, level LogLevel, context LogContextInterface) interface{} {
	return context.CallTime().UnixNano()
}

func verbn(message string, level LogLevel, context LogContextInterface) interface{} {
	return "\n"
}

func verbt(message string, level LogLevel, context LogContextInterface) interface{} {
	return "\t"
}

func createDateTimeVerbFunc(dateTimeFormat string) verbFunc {
	format := dateTimeFormat
	if format == "" {
		format = DateDefaultFormat
	}
	return func(message string, level LogLevel, context LogContextInterface) interface{} {
		return time.Now().Format(format)
	}
}

func createANSIEscapeFunc(escapeCodeString string) verbFunc {
	return func(message string, level LogLevel, context LogContextInterface) interface{} {
		if len(escapeCodeString) == 0 {
			return wrongEscapeCode
		}

		return fmt.Sprintf("%c[%sm", 0x1B, escapeCodeString)
	}
}
