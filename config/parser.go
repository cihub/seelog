// Package config contains configuration functionality of sealog
package config

import (
	"os"
	"io"
	. "sealog/common"
	"sealog/dispatchers"
	"sealog/writers"
	"strings"
	"fmt"
)

type elementMapEntry struct {
	constructor func(node *xmlNode, formatFromParent *formatDummy, formats map[string]*formatDummy) (interface{}, os.Error)
}

var elementMap map[string]elementMapEntry

func init() {
	elementMap = map[string]elementMapEntry{
		"outputs": {createSplitter},
		"file":    {createFileWriter},
		//"smtp": { createSmtpWriter },
	}
}

type formatDummy struct {
	id     string
	format string
}

// Creates a new config from given reader. Returns error if format is incorrect or anything happened.
func ConfigFromReader(reader io.Reader) (*LogConfig, os.Error) {
	config, err := unmarshalConfig(reader)
	if err != nil {
		return nil, err
	}

	if config.name != "sealog" {
		return nil, os.NewError("Root xml tag must be 'sealog'")
	}

	err = checkUnexpectedAttribute(config, "minlevel", "maxlevel", "levels")
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

	newConfig := &LogConfig{Constraints: constraints, Exceptions: exceptions, RootDispatcher: dispatcher}
	return newConfig, nil
}

func getConstraints(node *xmlNode) (LogLevelConstraints, os.Error) {
	minLevelStr, isMinLevel := node.attributes["minlevel"]
	maxLevelStr, isMaxLevel := node.attributes["maxlevel"]
	levelsStr, isLevels := node.attributes["levels"]

	if isLevels && (isMinLevel && isMaxLevel) {
		return nil, os.NewError("For level declaration use 'levels' OR 'minlevel', 'maxlevel'")
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
			return nil, os.NewError("Declared minlevel not found: " + minLevelStr)
		}
	}

	var maxLevel LogLevel = CriticalLvl
	if isMaxLevel {
		found := true
		maxLevel, found = LogLevelFromString(maxLevelStr)
		if !found {
			return nil, os.NewError("Declared maxlevel not found: " + maxLevelStr)
		}
	}

	return NewMinMaxConstraints(minLevel, maxLevel)
}

func getExceptions(config *xmlNode) ([]*LogLevelException, os.Error) {
	exceptions := make([]*LogLevelException, 0)

	var exceptionsNode *xmlNode
	for _, child := range config.children {
		if child.name == "exceptions" {
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
		if exceptionNode.name != "exception" {
			return nil, os.NewError("Incorrect nested element in exceptions section: " + exceptionNode.name)
		}

		err := checkUnexpectedAttribute(exceptionNode, "minlevel", "maxlevel", "levels", "funcpattern", "filepattern")
		if err != nil {
			return nil, err
		}

		constraints, err := getConstraints(exceptionNode)
		if err != nil {
			return nil, os.NewError("Incorrect exception node: " + err.String())
		}

		funcPattern, isFuncPattern := exceptionNode.attributes["funcpattern"]
		filePattern, isFilePattern := exceptionNode.attributes["filepattern"]
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

func getFormats(config *xmlNode) (map[string]*formatDummy, os.Error) {
	formats := make(map[string]*formatDummy, 0)

	var formatsNode *xmlNode
	for _, child := range config.children {
		if child.name == "formats" {
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
		if formatNode.name != "format" {
			return nil, os.NewError("Incorrect nested element in formats section: " + formatNode.name)
		}

		err := checkUnexpectedAttribute(formatNode, "id", "format")
		if err != nil {
			return nil, err
		}

		id, isId := formatNode.attributes["id"]
		formatStr, isFormat := formatNode.attributes["format"]
		if !isId {
			return nil, os.NewError("Format has no 'id' attribute")
		}
		if !isFormat {
			return nil, os.NewError("Format[" + id + "] has no 'format' attribute")
		}

		formats[id] = &formatDummy{id, formatStr}
	}

	return formats, nil
}

func getOutputsTree(config *xmlNode, formats map[string]*formatDummy) (dispatchers.DispatcherInterface, os.Error) {
	var outputsNode *xmlNode
	for _, child := range config.children {
		if child.name == "outputs" {
			outputsNode = child
			break
		}
	}

	if outputsNode != nil {
		err := checkUnexpectedAttribute(outputsNode, "formatid")
		if err != nil {
			return nil, err
		}

		format, err := getCurrentFormat(outputsNode, nil, formats)
		if err != nil {
			return nil, err
		}

		output, err := elementMap["outputs"].constructor(outputsNode, format, formats)
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
	return dispatchers.NewSplitDispatcher([]interface{}{console})
}

func getCurrentFormat(node *xmlNode, formatFromParent *formatDummy, formats map[string]*formatDummy) (*formatDummy, os.Error) {
	formatId, isFormatId := node.attributes["formatid"]
	if isFormatId {
		format, ok := formats[formatId]
		if !ok {
			return nil, os.NewError("Formatid = '" + formatId + "' doesn't exist")
		}

		return format, nil
	}

	return formatFromParent, nil
}

func createOutputs(node *xmlNode, format *formatDummy, formats map[string]*formatDummy) ([]interface{}, os.Error) {
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

func createSplitter(node *xmlNode, formatFromParent *formatDummy, formats map[string]*formatDummy) (interface{}, os.Error) {
	err := checkUnexpectedAttribute(node, "formatid")
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

	return dispatchers.NewSplitDispatcher(outputs)
}

func createFileWriter(node *xmlNode, formatFromParent *formatDummy, formats map[string]*formatDummy) (interface{}, os.Error) {
	err := checkUnexpectedAttribute(node, "formatid", "path")
	if err != nil {
		return nil, err
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}
	_ = currentFormat

	if len(node.children) > 0 {
		return nil, os.NewError("Output '" + node.name + "' must not have children")
	}

	path, isPath := node.attributes["path"]
	if !isPath {
		return nil, os.NewError("Output '" + node.name + "' has no 'path' attribute")
	}

	return writers.NewFileWriter(path)
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
