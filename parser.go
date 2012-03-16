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

package seelog

import (
	"time"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	SeelogConfigId             = "seelog"
	OutputsId                  = "outputs"
	FormatsId                  = "formats"
	MinLevelId                 = "minlevel"
	MaxLevelId                 = "maxlevel"
	LevelsId                   = "levels"
	ExceptionsId               = "exceptions"
	ExceptionId                = "exception"
	FuncPatternId              = "funcpattern"
	FilePatternId              = "filepattern"
	FormatId                   = "format"
	FormatAttrId               = "format"
	FormatKeyAttrId            = "id"
	OutputFormatId             = "formatid"
	FilePathId                 = "path"
	FileWriterId               = "file"
	SmtpWriterId               = "smtp"
	SenderAddressId            = "senderaddress"
	SenderNameId               = "sendername"
	RecipientId                = "recipient"
	AddressId                  = "address"
	HostNameId                 = "hostname"
	HostPortId                 = "hostport"
	UserNameId                 = "username"
	UserPassId                 = "password"
	SpliterDispatcherId        = "splitter"
	ConsoleWriterId            = "console"
	FilterDispatcherId         = "filter"
	FilterLevelsAttrId         = "levels"
	RollingFileWriterId        = "rollingfile"
	RollingFileTypeAttr        = "type"
	RollingFilePathAttr        = "filename"
	RollingFileMaxSizeAttr     = "maxsize"
	RollingFileMaxRollsAttr    = "maxrolls"
	RollingFileDataPatternAttr = "datepattern"
	bufferedWriterId           = "buffered"
	BufferedSizeAttr           = "size"
	BufferedFlushPeriodAttr    = "flushperiod"
	LoggerTypeFromStringAttr   = "type"
	AsyncLoggerIntervalAttr    = "asyncinterval"
)

var (
	nodeMustHaveChildrenError = errors.New("Node must have children")
	nodeCannotHaveChildrenError = errors.New("Node cannot have children")
)

type unexpectedChildElementError string

func (e unexpectedChildElementError) Error() string {
	return fmt.Sprintf("Unexpected child element: %s", e)
}

type elementMapEntry struct {
	constructor func(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error)
}

var elementMap map[string]elementMapEntry

func init() {
	elementMap = map[string]elementMapEntry{
		FileWriterId:        {createfileWriter},
		SpliterDispatcherId: {createSplitter},
		FilterDispatcherId:  {createFilter},
		ConsoleWriterId:     {createconsoleWriter},
		RollingFileWriterId: {createrollingFileWriter},
		bufferedWriterId:    {createbufferedWriter},
		SmtpWriterId: 		 {createsmtpWriter},
	}
}

// configFromReader parses data from a given reader. 
// Returns parsed config which can be used to create logger in case no errors occured.
// Returns error if format is incorrect or anything happened.
func configFromReader(reader io.Reader) (*logConfig, error) {
	config, err := unmarshalConfig(reader)
	if err != nil {
		return nil, err
	}

	if config.name != SeelogConfigId {
		return nil, errors.New("Root xml tag must be '" + SeelogConfigId + "'")
	}

	err = checkUnexpectedAttribute(config, MinLevelId, MaxLevelId, LevelsId, LoggerTypeFromStringAttr, AsyncLoggerIntervalAttr)
	if err != nil {
		return nil, err
	}

	err = checkExpectedElements(config, optionalElement(OutputsId), optionalElement(FormatsId), 
								optionalElement(ExceptionsId))
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

	loggerType, logData, err := getloggerTypeFromStringData(config)
	if err != nil {
		return nil, err
	}

	return newConfig(constraints, exceptions, dispatcher, loggerType, logData)
}

func getConstraints(node *xmlNode) (logLevelConstraints, error) {
	minLevelStr, isMinLevel := node.attributes[MinLevelId]
	maxLevelStr, isMaxLevel := node.attributes[MaxLevelId]
	levelsStr, isLevels := node.attributes[LevelsId]

	if isLevels && (isMinLevel && isMaxLevel) {
		return nil, errors.New("For level declaration use '" + LevelsId + "'' OR '" + MinLevelId +
			"', '" + MaxLevelId + "'")
	}

	offString := LogLevel(Off).String()

	if (isLevels && strings.TrimSpace(levelsStr) == offString) ||
		(isMinLevel && !isMaxLevel && minLevelStr == offString) {

		return newOffConstraints()
	}

	if isLevels {
		levels, err := parseLevels(levelsStr)
		if err != nil {
			return nil, err
		}
		return newListConstraints(levels)
	}

	var minLevel LogLevel = TraceLvl
	if isMinLevel {
		found := true
		minLevel, found = LogLevelFromString(minLevelStr)
		if !found {
			return nil, errors.New("Declared " + MinLevelId + " not found: " + minLevelStr)
		}
	}

	var maxLevel LogLevel = CriticalLvl
	if isMaxLevel {
		found := true
		maxLevel, found = LogLevelFromString(maxLevelStr)
		if !found {
			return nil, errors.New("Declared " + MaxLevelId + " not found: " + maxLevelStr)
		}
	}

	return newMinMaxConstraints(minLevel, maxLevel)
}

func parseLevels(str string) ([]LogLevel, error) {
	levelsStrArr := strings.Split(strings.Replace(str, " ", "", -1), ",")
	levels := make([]LogLevel, 0)
	for _, levelStr := range levelsStrArr {
		level, found := LogLevelFromString(levelStr)
		if !found {
			return nil, errors.New("Declared level not found: " + levelStr)
		}

		levels = append(levels, level)
	}

	return levels, nil
}

func getExceptions(config *xmlNode) ([]*logLevelException, error) {
	exceptions := make([]*logLevelException, 0)

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

	err = checkExpectedElements(exceptionsNode, multipleMandatoryElements("exception"))
	if err != nil {
		return nil, err
	}

	for _, exceptionNode := range exceptionsNode.children {
		if exceptionNode.name != ExceptionId {
			return nil, errors.New("Incorrect nested element in exceptions section: " + exceptionNode.name)
		}

		err := checkUnexpectedAttribute(exceptionNode, MinLevelId, MaxLevelId, LevelsId, FuncPatternId, FilePatternId)
		if err != nil {
			return nil, err
		}

		constraints, err := getConstraints(exceptionNode)
		if err != nil {
			return nil, errors.New("Incorrect " + ExceptionsId + " node: " + err.Error())
		}

		funcPattern, isFuncPattern := exceptionNode.attributes[FuncPatternId]
		filePattern, isFilePattern := exceptionNode.attributes[FilePatternId]
		if !isFuncPattern {
			funcPattern = "*"
		}
		if !isFilePattern {
			filePattern = "*"
		}

		exception, err := newLogLevelException(funcPattern, filePattern, constraints)
		if err != nil {
			return nil, errors.New("Incorrect exception node: " + err.Error())
		}

		exceptions = append(exceptions, exception)
	}

	return exceptions, nil
}

func checkDistinctExceptions(exceptions []*logLevelException) error {
	for i, exception := range exceptions {
		for j, exception1 := range exceptions {
			if i == j {
				continue
			}

			if exception.FuncPattern() == exception1.FuncPattern() &&
				exception.FilePattern() == exception1.FilePattern() {

				return errors.New(fmt.Sprintf("There are two or more duplicate exceptions. Func: %v, file% %v",
					exception.FuncPattern(), exception.FilePattern()))
			}
		}
	}

	return nil
}

func getFormats(config *xmlNode) (map[string]*formatter, error) {
	formats := make(map[string]*formatter, 0)

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

	err = checkExpectedElements(formatsNode, multipleMandatoryElements("format"))
	if err != nil {
		return nil, err
	}

	for _, formatNode := range formatsNode.children {
		if formatNode.name != FormatId {
			return nil, errors.New("Incorrect nested element in " + FormatsId + " section: " + formatNode.name)
		}

		err := checkUnexpectedAttribute(formatNode, FormatKeyAttrId, FormatId)
		if err != nil {
			return nil, err
		}

		id, isId := formatNode.attributes[FormatKeyAttrId]
		formatStr, isFormat := formatNode.attributes[FormatAttrId]
		if !isId {
			return nil, errors.New("Format has no '" + FormatKeyAttrId + "' attribute")
		}
		if !isFormat {
			return nil, errors.New("Format[" + id + "] has no '" + FormatAttrId + "' attribute")
		}

		formatter, err := newFormatter(formatStr)
		if err != nil {
			return nil, err
		}

		formats[id] = formatter
	}

	return formats, nil
}

func getloggerTypeFromStringData(config *xmlNode) (logType loggerTypeFromString, logData interface{}, err error) {
	logTypeStr, loggerTypeExists := config.attributes[LoggerTypeFromStringAttr]
	
	if !loggerTypeExists {
		return DefaultloggerTypeFromString, nil, nil
	}
	
	logType, found := loggerTypeFromStringFromString(logTypeStr)
	
	if !found {
		return 0, nil, errors.New(fmt.Sprintf("Unknown logger type: %s", logTypeStr))
	}
	
	if logType == asyncTimerloggerTypeFromString {
		intervalStr, intervalExists := config.attributes[AsyncLoggerIntervalAttr]
		if !intervalExists {
			return 0, nil, missingArgumentError(config.name, AsyncLoggerIntervalAttr)
		}

		interval, err := strconv.ParseUint(intervalStr, 10, 32)
		if err != nil {
			return 0, nil, err
		}
		
		logData = asyncTimerLoggerData{uint32(interval)}
	}
	
	return logType, logData, nil
}

func getOutputsTree(config *xmlNode, formats map[string]*formatter) (dispatcherInterface, error) {
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

		formatter, err := getCurrentFormat(outputsNode, Defaultformatter, formats)
		if err != nil {
			return nil, err
		}

		output, err := createSplitter(outputsNode, formatter, formats)
		if err != nil {
			return nil, err
		}

		dispatcher, ok := output.(dispatcherInterface)
		if ok {
			return dispatcher, nil
		}
	}

	console, err := newConsoleWriter()
	if err != nil {
		return nil, err
	}
	return newSplitDispatcher(Defaultformatter, []interface{}{console})
}

func getCurrentFormat(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (*formatter, error) {
	formatId, isFormatId := node.attributes[OutputFormatId]
	if isFormatId {
		format, ok := formats[formatId]
		if !ok {
			return nil, errors.New("Formatid = '" + formatId + "' doesn't exist")
		}

		return format, nil
	}

	return formatFromParent, nil
}

func createInnerReceivers(node *xmlNode, format *formatter, formats map[string]*formatter) ([]interface{}, error) {
	outputs := make([]interface{}, 0)
	for _, childNode := range node.children {
		entry, ok := elementMap[childNode.name]
		if !ok {
			return nil, errors.New("Unnknown tag '" + childNode.name + "' in outputs section")
		}

		output, err := entry.constructor(childNode, format, formats)
		if err != nil {
			return nil, err
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

func createSplitter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
	err := checkUnexpectedAttribute(node, OutputFormatId)
	if err != nil {
		return nil, err
	}

	if !node.hasChildren() {
		return nil, nodeMustHaveChildrenError
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	receivers, err := createInnerReceivers(node, currentFormat, formats)
	if err != nil {
		return nil, err
	}

	return newSplitDispatcher(currentFormat, receivers)
}

func createFilter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
	err := checkUnexpectedAttribute(node, OutputFormatId, FilterLevelsAttrId)
	if err != nil {
		return nil, err
	}

	if !node.hasChildren() {
		return nil, nodeMustHaveChildrenError
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	levelsStr, isLevels := node.attributes[FilterLevelsAttrId]
	if !isLevels {
		return nil, missingArgumentError(node.name, FilterLevelsAttrId)
	}

	levels, err := parseLevels(levelsStr)
	if err != nil {
		return nil, err
	}

	receivers, err := createInnerReceivers(node, currentFormat, formats)
	if err != nil {
		return nil, err
	}

	return newFilterDispatcher(currentFormat, receivers, levels...)
}

func createfileWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
	err := checkUnexpectedAttribute(node, OutputFormatId, FilePathId)
	if err != nil {
		return nil, err
	}

	if node.hasChildren() {
		return nil, nodeCannotHaveChildrenError
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	path, isPath := node.attributes[FilePathId]
	if !isPath {
		return nil, missingArgumentError(node.name, FilePathId)
	}

	fileWriter, err := newFileWriter(path)
	if err != nil {
		return nil, err
	}

	return newFormattedWriter(fileWriter, currentFormat)
}

// Creates new SMTP writer if encountered in the config file
func createsmtpWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
	err := checkUnexpectedAttribute(
		node,
		OutputFormatId,
		SenderAddressId,
		SenderNameId,
		//RecipientsId,
		HostNameId,
		HostPortId,
		UserNameId,
		UserPassId,
	)
	if err != nil {
		return nil, err
	}
	if !node.hasChildren() {
		return nil, nodeMustHaveChildrenError
	}
	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}
	senderAddress, ok := node.attributes[SenderAddressId]
	if !ok {
		return nil, missingArgumentError(node.name, SenderAddressId)
	}
	senderName, ok := node.attributes[SenderNameId]
	if !ok {
		return nil, missingArgumentError(node.name, SenderNameId)
	}
	recipientAddresses := make([]string, len(node.children))
	// Extract recipient addresses from child nodes
	for i, childNode := range node.children {
		if childNode.name == RecipientId {
			address, ok := childNode.attributes[AddressId]
			if !ok {
				return nil, missingArgumentError(childNode.name, AddressId)
			}
			recipientAddresses[i] = address
		} else {
			return nil, unexpectedChildElementError(childNode.name)
		}
	}
	hostName, ok := node.attributes[HostNameId]
	if !ok {
		return nil, missingArgumentError(node.name, HostNameId)
	}
	hostPort, ok := node.attributes[HostPortId]
	if !ok {
		return nil, missingArgumentError(node.name, HostPortId)
	}
	// Get int value from string
	hPort, err := strconv.Atoi(hostPort)
	if err != nil {
		return nil, errors.New("Invalid host port number")
	}
	userName, ok := node.attributes[UserNameId]
	if !ok {
		return nil, missingArgumentError(node.name, UserNameId)
	}
	userPass, ok := node.attributes[UserPassId]
	if !ok {
		return nil, missingArgumentError(node.name, UserPassId)
	}
	smtpWriter, err := newSmtpWriter(
		senderAddress,
		senderName,
		recipientAddresses,
		hostName,
		hPort,
		userName,
		userPass,
	)
	if err != nil {
		return nil, err
	}
	return newFormattedWriter(smtpWriter, currentFormat)
}

func createconsoleWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
	err := checkUnexpectedAttribute(node, OutputFormatId)
	if err != nil {
		return nil, err
	}

	if node.hasChildren() {
		return nil, nodeCannotHaveChildrenError
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	consoleWriter, err := newConsoleWriter()
	if err != nil {
		return nil, err
	}

	return newFormattedWriter(consoleWriter, currentFormat)
}

func createrollingFileWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
	if node.hasChildren() {
		return nil, nodeCannotHaveChildrenError
	}

	rollingTypeStr, isRollingType := node.attributes[RollingFileTypeAttr]
	if !isRollingType {
		return nil, missingArgumentError(node.name, RollingFileTypeAttr)
	}

	rollingType, ok := rollingTypeFromString(rollingTypeStr)
	if !ok {
		return nil, errors.New("Unknown rolling file type: " + rollingTypeStr)
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	path, isPath := node.attributes[RollingFilePathAttr]
	if !isPath {
		return nil, missingArgumentError(node.name, RollingFilePathAttr)
	}

	if rollingType == Size {
		err := checkUnexpectedAttribute(node, OutputFormatId, RollingFileTypeAttr, RollingFilePathAttr, RollingFileMaxSizeAttr, RollingFileMaxRollsAttr)
		if err != nil {
			return nil, err
		}

		maxSizeStr, isMaxSize := node.attributes[RollingFileMaxSizeAttr]
		if !isMaxSize {
			return nil, missingArgumentError(node.name, RollingFileMaxSizeAttr)
		}

		maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
		if err != nil {
			return nil, err
		}

		maxRollsStr, isMaxRolls := node.attributes[RollingFileMaxRollsAttr]
		if !isMaxRolls {
			return nil, missingArgumentError(node.name, RollingFileMaxRollsAttr)
		}

		maxRolls, err := strconv.Atoi(maxRollsStr)
		if err != nil {
			return nil, err
		}

		rollingWriter, err := newRollingFileWriterSize(path, maxSize, maxRolls)
		if err != nil {
			return nil, err
		}

		return newFormattedWriter(rollingWriter, currentFormat)

	} else if rollingType == Date {
		err := checkUnexpectedAttribute(node, OutputFormatId, RollingFileTypeAttr, RollingFilePathAttr, RollingFileDataPatternAttr)
		if err != nil {
			return nil, err
		}

		dataPattern, isDataPattern := node.attributes[RollingFileDataPatternAttr]
		if !isDataPattern {
			return nil, missingArgumentError(node.name, RollingFileDataPatternAttr)
		}

		rollingWriter, err := newRollingFileWriterDate(path, dataPattern)
		if err != nil {
			return nil, err
		}

		return newFormattedWriter(rollingWriter, currentFormat)
	}

	return nil, errors.New("Incorrect rolling writer type " + rollingTypeStr)
}

func createbufferedWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
	err := checkUnexpectedAttribute(node, OutputFormatId, BufferedSizeAttr, BufferedFlushPeriodAttr)
	if err != nil {
		return nil, err
	}

	if !node.hasChildren() {
		return nil, nodeMustHaveChildrenError
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	sizeStr, isSize := node.attributes[BufferedSizeAttr]
	if !isSize {
		return nil, missingArgumentError(node.name, BufferedSizeAttr)
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return nil, err
	}

	flushPeriod := 0
	flushPeriodStr, isFlushPeriod := node.attributes[BufferedFlushPeriodAttr]
	if isFlushPeriod {
		flushPeriod, err = strconv.Atoi(flushPeriodStr)
		if err != nil {
			return nil, err
		}
	}

	// Inner writer couldn't have its own format, so we pass 'currentFormat' as its parent format
	receivers, err := createInnerReceivers(node, currentFormat, formats)
	if err != nil {
		return nil, err
	}

	formattedWriter, ok := receivers[0].(*formattedWriter)
	if !ok {
		return nil, errors.New("Buffered writer's child is not writer")
	}

	// ... and then we check that it hasn't changed
	if formattedWriter.Format() != currentFormat {
		return nil, errors.New("Inner writer can not have his own format")
	}

	bufferedWriter, err := newBufferedWriter(formattedWriter.Writer(), size, time.Duration(flushPeriod))
	if err != nil {
		return nil, err
	}

	return newFormattedWriter(bufferedWriter, currentFormat)
}

func checkUnexpectedAttribute(node *xmlNode, expectedAttrs ...string) error {
	for attr, _ := range node.attributes {
		isExpected := false
		for _, expected := range expectedAttrs {
			if attr == expected {
				isExpected = true
				break
			}
		}
		if !isExpected {
			return errors.New(node.name + " has unexpected attribute: " + attr)
		}
	}

	return nil
}

type expectedElementInfo struct {
	name      string
	mandatory bool
	multiple  bool
}

func optionalElement(name string) expectedElementInfo {
	return expectedElementInfo{name, false, false}
}
func mandatoryElement(name string) expectedElementInfo {
	return expectedElementInfo{name, true, false}
}
func multipleElements(name string) expectedElementInfo {
	return expectedElementInfo{name, false, true}
}
func multipleMandatoryElements(name string) expectedElementInfo {
	return expectedElementInfo{name, true, true}
}


func checkExpectedElements(node *xmlNode, elements ...expectedElementInfo) error {
	for _, element := range elements {
		count := 0
		for _, child := range node.children {
			if child.name == element.name {
				count++
			}
		}

		if count == 0 && element.mandatory {
			return errors.New(node.name + " does not have mandatory subnode - " + element.name)
		}
		if count > 1 && !element.multiple {
			return errors.New(node.name + " has more then one subnode - " + element.name)
		}
	}

	for _, child := range node.children {
		isExpected := false
		for _, element := range elements {
			if child.name == element.name {
				isExpected = true
			}
		}

		if !isExpected {
			return errors.New(node.name + " has unexpected child: " + child.name)
		}
	}

	return nil
}

func missingArgumentError(nodeName string, attrName string) error {
	return errors.New("Output '" + nodeName + "' has no '" + attrName + "' attribute")
}
