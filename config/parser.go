// Package config contains configuration functionality of sealog.
package config

import (
	"os"
	"io"
	. "github.com/cihub/sealog/common"
	"github.com/cihub/sealog/dispatchers"
	"github.com/cihub/sealog/writers"
	"github.com/cihub/sealog/format"
	"strings"
	"fmt"
)

const (
	SealogConfigId = "sealog"
	OutputsId = "outputs"
	FormatsId = "formats"
	MinLevelId = "minlevel"
	MaxLevelId = "maxlevel"
	LevelsId = "levels"
	ExceptionsId = "exceptions"
	ExceptionId = "exception"
	FuncPatternId = "funcpattern"
	FilePatternId = "filepattern"
	FormatId = "format"
	FormatAttrId = "format"
	FormatKeyAttrId = "id"
	OutputFormatId = "formatid"
	FilePathId = "path"
	FileWriterId = "file"
	SpliterDispatcherId = "splitter"
)

type elementMapEntry struct {
	constructor func(node *xmlNode, formatFromParent *format.Formatter, formats map[string]*format.Formatter) (interface{}, os.Error)
}

var elementMap map[string]elementMapEntry

func init() {
	elementMap = map[string]elementMapEntry{
		OutputsId: {createSplitter},
		FileWriterId:    {createFileWriter},
		SpliterDispatcherId: {createSplitter},
		//"smtp": { createSmtpWriter },
	}
}

// Creates a new config from given reader. Returns error if format is incorrect or anything happened.
func ConfigFromReader(reader io.Reader) (*LogConfig, os.Error) {
	config, err := unmarshalConfig(reader)
	if err != nil {
		return nil, err
	}

	if config.name != SealogConfigId {
		return nil, os.NewError("Root xml tag must be '" + SealogConfigId + "'")
	}

	err = checkUnexpectedAttribute(config, MinLevelId, MaxLevelId, LevelsId)
	if err != nil {
		return nil, err
	}

	constraints, err := getConstraints(config)
	if err != nil {
		return nil, err
	}

	exceptions, err := getExceptions(config)
	if err != nil {
		return nil, err
	}
	err = checkDistinctExceptions(exceptions)
	if err != nil {
		return nil, err
	}

	formats, err := getFormats(config)
	if err != nil {
		return nil, err
	}

	dispatcher, err := getOutputsTree(config, formats)
	if err != nil {
		return nil, err
	}

	return NewConfig(constraints, exceptions, dispatcher)
}

func getConstraints(node *xmlNode) (LogLevelConstraints, os.Error) {
	minLevelStr, isMinLevel := node.attributes[MinLevelId]
	maxLevelStr, isMaxLevel := node.attributes[MaxLevelId]
	levelsStr, isLevels := node.attributes[LevelsId]

	if isLevels && (isMinLevel && isMaxLevel) {
		return nil, os.NewError("For level declaration use '" + LevelsId + "'' OR '" + MinLevelId + 
			"', '" + MaxLevelId + "'")
	}

	offString := LogLevel(Off).String()

	if (isLevels && strings.TrimSpace(levelsStr) == offString) ||
		(isMinLevel && !isMaxLevel && minLevelStr == offString) {

		return NewOffConstraints()
	}

	if isLevels {
		levelsStrArr := strings.Split(strings.Replace(levelsStr, " ", "", -1), ",")
		levels := make([]LogLevel, 0)
		for _, levelStr := range levelsStrArr {
			level, found := LogLevelFromString(levelStr)
			if !found {
				return nil, os.NewError("Declared level not found: " + levelStr)
			}

			levels = append(levels, level)
		}
		return NewListConstraints(levels)
	}

	var minLevel LogLevel = TraceLvl
	if isMinLevel {
		found := true
		minLevel, found = LogLevelFromString(minLevelStr)
		if !found {
			return nil, os.NewError("Declared " + MinLevelId + " not found: " + minLevelStr)
		}
	}

	var maxLevel LogLevel = CriticalLvl
	if isMaxLevel {
		found := true
		maxLevel, found = LogLevelFromString(maxLevelStr)
		if !found {
			return nil, os.NewError("Declared " + MaxLevelId + " not found: " + maxLevelStr)
		}
	}

	return NewMinMaxConstraints(minLevel, maxLevel)
}

func getExceptions(config *xmlNode) ([]*LogLevelException, os.Error) {
	exceptions := make([]*LogLevelException, 0)

	var exceptionsNode *xmlNode
	for _, child := range config.children {
		if child.name == ExceptionsId {
			exceptionsNode = child
			break
		}
	}

	if exceptionsNode == nil {
		return exceptions, nil
	}

	err := checkUnexpectedAttribute(exceptionsNode)
	if err != nil {
		return nil, err
	}

	for _, exceptionNode := range exceptionsNode.children {
		if exceptionNode.name != ExceptionId {
			return nil, os.NewError("Incorrect nested element in exceptions section: " + exceptionNode.name)
		}

		err := checkUnexpectedAttribute(exceptionNode, MinLevelId, MaxLevelId, LevelsId, FuncPatternId, FilePatternId)
		if err != nil {
			return nil, err
		}

		constraints, err := getConstraints(exceptionNode)
		if err != nil {
			return nil, os.NewError("Incorrect " + ExceptionsId + " node: " + err.String())
		}

		funcPattern, isFuncPattern := exceptionNode.attributes[FuncPatternId]
		filePattern, isFilePattern := exceptionNode.attributes[FilePatternId]
		if !isFuncPattern {
			funcPattern = "*"
		}
		if !isFilePattern {
			filePattern = "*"
		}

		exception, err := NewLogLevelException(funcPattern, filePattern, constraints)
		if err != nil {
			return nil, os.NewError("Incorrect exception node: " + err.String())
		}

		exceptions = append(exceptions, exception)
	}

	return exceptions, nil
}

func checkDistinctExceptions(exceptions []*LogLevelException) os.Error {
	for i, exception := range exceptions {
		for j, exception1 := range exceptions {
			if i == j {
				continue
			}

			if exception.FuncPattern() == exception1.FuncPattern() &&
				exception.FilePattern() == exception1.FilePattern() {

				return os.NewError(fmt.Sprintf("There are two or more duplicate exceptions. Func: %v, file% %v",
					exception.FuncPattern(), exception.FilePattern()))
			}
		}
	}

	return nil
}

func getFormats(config *xmlNode) (map[string]*format.Formatter, os.Error) {
	formats := make(map[string]*format.Formatter, 0)

	var formatsNode *xmlNode
	for _, child := range config.children {
		if child.name == FormatsId {
			formatsNode = child
			break
		}
	}

	if formatsNode == nil {
		return formats, nil
	}

	err := checkUnexpectedAttribute(formatsNode)
	if err != nil {
		return nil, err
	}

	for _, formatNode := range formatsNode.children {
		if formatNode.name != FormatId {
			return nil, os.NewError("Incorrect nested element in " + FormatsId + " section: " + formatNode.name)
		}

		err := checkUnexpectedAttribute(formatNode, FormatKeyAttrId, FormatId)
		if err != nil {
			return nil, err
		}

		id, isId := formatNode.attributes[FormatKeyAttrId]
		formatStr, isFormat := formatNode.attributes[FormatAttrId]
		if !isId {
			return nil, os.NewError("Format has no '" + FormatKeyAttrId + "' attribute")
		}
		if !isFormat {
			return nil, os.NewError("Format[" + id + "] has no '" + FormatAttrId + "' attribute")
		}

		formatter, err := format.NewFormatter(formatStr)
		if err != nil {
			return nil, err
		}
		
		formats[id] = formatter
	}

	return formats, nil
}

func getOutputsTree(config *xmlNode, formats map[string]*format.Formatter) (dispatchers.DispatcherInterface, os.Error) {
	var outputsNode *xmlNode
	for _, child := range config.children {
		if child.name == OutputsId {
			outputsNode = child
			break
		}
	}

	if outputsNode != nil {
		err := checkUnexpectedAttribute(outputsNode, OutputFormatId)
		if err != nil {
			return nil, err
		}

		formatter, err := getCurrentFormat(outputsNode, format.DefaultFormatter, formats)
		if err != nil {
			return nil, err
		}

		output, err := elementMap[OutputsId].constructor(outputsNode, formatter, formats)
		if err != nil {
			return nil, err
		}

		dispatcher, ok := output.(dispatchers.DispatcherInterface)
		if ok {
			return dispatcher, nil
		}
	}

	console, err := writers.NewConsoleWriter()
	if err != nil {
		return nil, err
	}
	return dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{console})
}

func getCurrentFormat(node *xmlNode, formatFromParent *format.Formatter, formats map[string]*format.Formatter) (*format.Formatter, os.Error) {
	formatId, isFormatId := node.attributes[OutputFormatId]
	if isFormatId {
		format, ok := formats[formatId]
		if !ok {
			return nil, os.NewError("Formatid = '" + formatId + "' doesn't exist")
		}

		return format, nil
	}

	return formatFromParent, nil
}

func createOutputs(node *xmlNode, format *format.Formatter, formats map[string]*format.Formatter) ([]interface{}, os.Error) {
	outputs := make([]interface{}, 0)
	for _, childNode := range node.children {
		entry, ok := elementMap[childNode.name]
		if !ok {
			return nil, os.NewError("Unnknown tag '" + childNode.name + "' in outputs section")
		}

		output, err := entry.constructor(childNode, format, formats)
		if err != nil {
			return nil, err
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

func createSplitter(node *xmlNode, formatFromParent *format.Formatter, formats map[string]*format.Formatter) (interface{}, os.Error) {
	err := checkUnexpectedAttribute(node, OutputFormatId)
	if err != nil {
		return nil, err
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	outputs, err := createOutputs(node, currentFormat, formats)
	if err != nil {
		return nil, err
	}

	return dispatchers.NewSplitDispatcher(currentFormat, outputs)
}

func createFileWriter(node *xmlNode, formatFromParent *format.Formatter, formats map[string]*format.Formatter) (interface{}, os.Error) {
	err := checkUnexpectedAttribute(node, OutputFormatId, FilePathId)
	if err != nil {
		return nil, err
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	if len(node.children) > 0 {
		return nil, os.NewError("Output '" + node.name + "' must not have children")
	}

	path, isPath := node.attributes[FilePathId]
	if !isPath {
		return nil, os.NewError("Output '" + node.name + "' has no '" + FilePathId + "' attribute")
	}

	fileWriter, err := writers.NewFileWriter(path)
	if err != nil {
		return nil, err
	}
	
	return dispatchers.NewFormattedWriter(fileWriter, currentFormat)
}

func checkUnexpectedAttribute(node *xmlNode, expectedAttrs ...string) os.Error {
	for attr, _ := range node.attributes {
		isExpected := false
		for _, expected := range expectedAttrs {
			if attr == expected {
				isExpected = true
				break
			}
		}
		if !isExpected {
			return os.NewError("Unexpected attribute: " + attr)
		}
	}

	return nil
}
