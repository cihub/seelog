// Package format contains formatting logic for sealog package and all available formats.
package format

import (
	"os"
	"fmt"
	"strings"
	"strconv"
	"unicode"
	"utf8"
	"time"
	. "sealog/common"
)

const (
	VerbSymbol = '%'
	VerbSymbolString = "%"
	VerbParameterStart = '('
	VerbParameterEnd = ')'
	DateDefaultFormat = "2006-01-02"
	TimeFormat = "15:04:05"
)

var DefaultFormatter *Formatter;

func init() {
	var err os.Error
	DefaultFormatter, err = NewFormatter("%Date/%Time [%Level] %Msg")
	if err != nil {
		fmt.Println("Error during DefaultFormatter creation: " + err.String())
	}
}

type verbFunc func(message string, level LogLevel, context *LogContext) string
type verbFuncCreator func(param string) verbFunc

var verbFuncs = map[string]verbFunc {
	"Level": verbLevel,
	"Lev": verbLev,
	"LEVEL": verbLEVEL,
	"LEV": verbLEV,
	"l": verbl,
	"Msg": verbMsg,
	"FullPath": verbFullPath,
	"File": verbFile,
	"RelFile": verbRelFile,
	"Func": verbFunction,
	"Time": verbTime,
}

var verbFuncsParametrized = map[string]verbFuncCreator {
	"Date": createDateTimeVerbFunc,
}

// Format is used to write messages in a specific format, inserting such additional data
// as log level, date/time, etc.
type Formatter struct {
	fmtStringOriginal string
	fmtString string
	verbFuncs []verbFunc
}

// NewFormatter creates a new formatter using a format string
func NewFormatter(formatString string) (*Formatter, os.Error) {
	newFormatter := new(Formatter)
	newFormatter.fmtStringOriginal = formatString
	
	err := newFormatter.buildVerbFuncs()
	if err != nil {
		//fmt.Println("Error: " + err.String())
		return nil, err
	}
	
	return newFormatter, nil
}

func (formatter *Formatter) buildVerbFuncs() os.Error {
	//fmt.Println("buildVerbFuncs for " + formatter.fmtStringOriginal)
	formatter.verbFuncs = make([]verbFunc, 0)
	var fmtString string 
	for i := 0; i < len(formatter.fmtStringOriginal); i++ {
		char := formatter.fmtStringOriginal[i]
		if char != VerbSymbol {
			fmtString += string(char)
			continue
		}
		
		isEndOfStr := i == len(formatter.fmtStringOriginal) - 1
		if isEndOfStr {
			return os.NewError(fmt.Sprintf("Format error: %v - last symbol", VerbSymbolString))
		}
		
		isDoubledVerbSymbol := formatter.fmtStringOriginal[i + 1] == VerbSymbol
		if isDoubledVerbSymbol {
			fmtString += VerbSymbolString
			i++
			continue
		}
		
		function, nextI, err := formatter.extractVerbFunc(i + 1)
		if err != nil {
			return err
		}
		
		fmtString += "%s"
		i = nextI
		//fmt.Println("Add func for " + verb)
		formatter.verbFuncs = append(formatter.verbFuncs, function)
	}
	
	//fmt.Println("FmtStr = " + fmtString)
	formatter.fmtString = fmtString
	return nil	
}

func (formatter *Formatter) extractVerbFunc(index int) (verbFunc, int, os.Error) {
	letterSequence := formatter.extractLetterSequence(index)
	if len(letterSequence) == 0 {
		return nil, 0, os.NewError(fmt.Sprintf("Format error: lack of verb after %v. At %v", VerbSymbolString, index))
	}
	
	function, verbLength, ok := formatter.findVerbFunc(letterSequence)
	if ok {
		return function, index + verbLength - 1, nil
	}
	
	function, verbLength, ok = formatter.findVerbFuncParametrized(letterSequence, index)
	if ok {
		return function, index + verbLength - 1, nil
	}
	
	return nil, 0, os.NewError("Format error: unrecognized verb at " + strconv.Itoa(index))
}

func (formatter *Formatter) extractLetterSequence(index int) string {
	letters := ""
	parsedStr := utf8.NewString(formatter.fmtStringOriginal[index:])
	for i := 0; i < parsedStr.RuneCount(); i++ {
		rune := parsedStr.At(i)
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
	for i:=0; i < len(letters); i++ {
		function, ok := verbFuncs[currentVerb]
		if ok {
			return function, len(currentVerb), ok
		}
		currentVerb = currentVerb[: len(currentVerb) - 1]
	}
	
	return nil, 0, false
}

func (formatter *Formatter) findVerbFuncParametrized(letters string, lettersStartIndex int) (verbFunc, int, bool) {
	currentVerb := letters
	for i:=0; i < len(letters); i++ {
		functionCreator, ok := verbFuncsParametrized[currentVerb]
		if ok {
			paramter := ""
			parameterLen := 0
			isVerbEqualsLetters := i == 0; // if not, then letter goes after verb, and verb is parameterless
			if isVerbEqualsLetters {
				userParamter := ""
				userParamter, parameterLen, ok = formatter.findparameter(lettersStartIndex + len(currentVerb))
				if ok {
					paramter = userParamter
				}
			}
			
			return functionCreator(paramter), len(currentVerb) + parameterLen, true
		}
		
		currentVerb = currentVerb[: len(currentVerb) - 1]
	}
	
	return nil, 0, false
}

func (formatter *Formatter) findparameter(startIndex int) (string, int, bool) {
	if len(formatter.fmtStringOriginal) == startIndex || formatter.fmtStringOriginal[startIndex] != VerbParameterStart {
		return "", 0, false
	}
	
	endIndex := strings.Index(formatter.fmtStringOriginal[startIndex: ], string(VerbParameterEnd)) + startIndex
	if endIndex == -1 {
		return "", 0, false
	}
	
	length := endIndex - startIndex + 1
	
	return formatter.fmtStringOriginal[startIndex + 1: endIndex], length, true
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

func verbLevel(message string, level LogLevel, context *LogContext) string  {
	levelStr, ok := levelToString[level]
	if !ok {
		return wrongLogLevel
	}
	return levelStr
}

func verbLev(message string, level LogLevel, context *LogContext) string  {
	levelStr, ok := levelToShortString[level]
	if !ok {
		return wrongLogLevel
	}
	return levelStr
}

func verbLEVEL(message string, level LogLevel, context *LogContext) string  {
	return strings.ToTitle(verbLevel(message, level, context))
}

func verbLEV(message string, level LogLevel, context *LogContext) string  {
	return strings.ToTitle(verbLev(message, level, context))
}

func verbl(message string, level LogLevel, context *LogContext) string  {
	levelStr, ok := levelToShortestString[level]
	if !ok {
		return wrongLogLevel
	}
	return levelStr
}

func verbMsg(message string, level LogLevel, context *LogContext) string  {
	return message
}

func verbFullPath(message string, level LogLevel, context *LogContext) string  {
	return context.FullPath()
}

func verbFile(message string, level LogLevel, context *LogContext) string  {
	return context.FileName()
}

func verbRelFile(message string, level LogLevel, context *LogContext) string  {
	return context.ShortPath()
}

func verbFunction(message string, level LogLevel, context *LogContext) string  {
	return context.Func()
}

func verbTime(message string, level LogLevel, context *LogContext) string  {
	return time.LocalTime().Format(TimeFormat)
}

func createDateTimeVerbFunc(dateTimeFormat string) verbFunc {
	format := dateTimeFormat
	if format == ""{
		format = DateDefaultFormat
	}
	return func(message string, level LogLevel, context *LogContext) string {
		return time.LocalTime().Format(format)
	}
}
