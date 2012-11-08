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
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	SeelogConfigId                  = "seelog"
	OutputsId                       = "outputs"
	FormatsId                       = "formats"
	MinLevelId                      = "minlevel"
	MaxLevelId                      = "maxlevel"
	LevelsId                        = "levels"
	ExceptionsId                    = "exceptions"
	ExceptionId                     = "exception"
	FuncPatternId                   = "funcpattern"
	FilePatternId                   = "filepattern"
	FormatId                        = "format"
	FormatAttrId                    = "format"
	FormatKeyAttrId                 = "id"
	OutputFormatId                  = "formatid"
	FilePathId                      = "path"
	FileWriterId                    = "file"
	SmtpWriterId                    = "smtp"
	SenderAddressId                 = "senderaddress"
	SenderNameId                    = "sendername"
	RecipientId                     = "recipient"
	AddressId                       = "address"
	HostNameId                      = "hostname"
	HostPortId                      = "hostport"
	UserNameId                      = "username"
	UserPassId                      = "password"
	CACertificatePathsId            = "cacertificatepaths"
	SpliterDispatcherId             = "splitter"
	ConsoleWriterId                 = "console"
	FilterDispatcherId              = "filter"
	FilterLevelsAttrId              = "levels"
	RollingFileWriterId             = "rollingfile"
	RollingFileTypeAttr             = "type"
	RollingFilePathAttr             = "filename"
	RollingFileMaxSizeAttr          = "maxsize"
	RollingFileMaxRollsAttr         = "maxrolls"
	RollingFileDataPatternAttr      = "datepattern"
	RollingFileArchiveAttr          = "archive"
	RollingFileArchivePathAttr      = "archivepath"
	bufferedWriterId                = "buffered"
	BufferedSizeAttr                = "size"
	BufferedFlushPeriodAttr         = "flushperiod"
	LoggerTypeFromStringAttr        = "type"
	AsyncLoggerIntervalAttr         = "asyncinterval"
	AdaptLoggerMinIntervalAttr      = "mininterval"
	AdaptLoggerMaxIntervalAttr      = "maxinterval"
	AdaptLoggerCriticalMsgCountAttr = "critmsgcount"
	PredefinedPrefix                = "std:"
	ConnWriterId                    = "conn"
	ConnWriterAddrAttr              = "addr"
	ConnWriterNetAttr               = "net"
	ConnWriterReconnectOnMsgAttr    = "reconnectonmsg"
)

type elementMapEntry struct {
	constructor func(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error)
}

var elementMap map[string]elementMapEntry
var predefinedFormats map[string]*formatter

func init() {
	elementMap = map[string]elementMapEntry{
		FileWriterId:        {createfileWriter},
		SpliterDispatcherId: {createSplitter},
		FilterDispatcherId:  {createFilter},
		ConsoleWriterId:     {createConsoleWriter},
		RollingFileWriterId: {createRollingFileWriter},
		bufferedWriterId:    {createbufferedWriter},
		SmtpWriterId:        {createSmtpWriter},
		ConnWriterId:        {createconnWriter},
	}

	err := fillPredefinedFormats()

	if err != nil {
		panic(fmt.Sprintf("Seelog couldn't start: predefined formats creation failed. Error: %s", err.Error()))
	}
}

func fillPredefinedFormats() error {
	predefinedFormatsWithoutPrefix := map[string]string{
		"xml-debug":       `<time>%Ns</time><lev>%Lev</lev><msg>%Msg</msg><path>%RelFile</path><func>%Func</func>`,
		"xml-debug-short": `<t>%Ns</t><l>%l</l><m>%Msg</m><p>%RelFile</p><f>%Func</f>`,
		"xml":             `<time>%Ns</time><lev>%Lev</lev><msg>%Msg</msg>`,
		"xml-short":       `<t>%Ns</t><l>%l</l><m>%Msg</m>`,

		"json-debug":       `{"time":%Ns,"lev":"%Lev","msg":"%Msg","path":"%RelFile","func":"%Func"}`,
		"json-debug-short": `{"t":%Ns,"l":"%Lev","m":"%Msg","p":"%RelFile","f":"%Func"}`,
		"json":             `{"time":%Ns,"lev":"%Lev","msg":"%Msg"}`,
		"json-short":       `{"t":%Ns,"l":"%Lev","m":"%Msg"}`,

		"debug":       `[%LEVEL] %RelFile:%Func %Date %Time %Msg%n`,
		"debug-short": `[%LEVEL] %Date %Time %Msg%n`,
		"fast":        `%Ns %l %Msg%n`,
	}

	predefinedFormats = make(map[string]*formatter)

	for formatKey, format := range predefinedFormatsWithoutPrefix {

		formatter, err := newFormatter(format)
		if err != nil {
			return err
		}

		predefinedFormats[PredefinedPrefix+formatKey] = formatter
	}

	return nil
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

	err = checkUnexpectedAttribute(config, MinLevelId, MaxLevelId, LevelsId, LoggerTypeFromStringAttr,
		AsyncLoggerIntervalAttr, AdaptLoggerMinIntervalAttr, AdaptLoggerMaxIntervalAttr,
		AdaptLoggerCriticalMsgCountAttr)
	if err != nil {
		return nil, err
	}

	err = checkExpectedElements(config, optionalElement(OutputsId), optionalElement(FormatsId), optionalElement(ExceptionsId))
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
		// If we open several files, but then fail to parse the config, we should close
		// those files before reporting that config is invalid.
		if dispatcher != nil {
			dispatcher.Close()
		}

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

	logType, found := getLoggerTypeFromString(logTypeStr)

	if !found {
		return 0, nil, errors.New(fmt.Sprintf("Unknown logger type: %s", logTypeStr))
	}

	if logType == asyncTimerloggerTypeFromString {
		intervalStr, intervalExists := config.attributes[AsyncLoggerIntervalAttr]
		if !intervalExists {
			return 0, nil, newMissingArgumentError(config.name, AsyncLoggerIntervalAttr)
		}

		interval, err := strconv.ParseUint(intervalStr, 10, 32)
		if err != nil {
			return 0, nil, err
		}

		logData = asyncTimerLoggerData{uint32(interval)}
	} else if logType == adaptiveLoggerTypeFromString {

		// Min interval
		minIntStr, minIntExists := config.attributes[AdaptLoggerMinIntervalAttr]
		if !minIntExists {
			return 0, nil, newMissingArgumentError(config.name, AdaptLoggerMinIntervalAttr)
		}
		minInterval, err := strconv.ParseUint(minIntStr, 10, 32)
		if err != nil {
			return 0, nil, err
		}

		// Max interval
		maxIntStr, maxIntExists := config.attributes[AdaptLoggerMaxIntervalAttr]
		if !maxIntExists {
			return 0, nil, newMissingArgumentError(config.name, AdaptLoggerMaxIntervalAttr)
		}
		maxInterval, err := strconv.ParseUint(maxIntStr, 10, 32)
		if err != nil {
			return 0, nil, err
		}

		// Critical msg count
		criticalMsgCountStr, criticalMsgCountExists := config.attributes[AdaptLoggerCriticalMsgCountAttr]
		if !criticalMsgCountExists {
			return 0, nil, newMissingArgumentError(config.name, AdaptLoggerCriticalMsgCountAttr)
		}
		criticalMsgCount, err := strconv.ParseUint(criticalMsgCountStr, 10, 32)
		if err != nil {
			return 0, nil, err
		}

		logData = adaptiveLoggerData{uint32(minInterval), uint32(maxInterval), uint32(criticalMsgCount)}
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
	if !isFormatId {
		return formatFromParent, nil
	}

	format, ok := formats[formatId]
	if ok {
		return format, nil
	}

	// Test for predefined format match
	pdFormat, pdOk := predefinedFormats[formatId]

	if !pdOk {
		return nil, errors.New("Formatid = '" + formatId + "' doesn't exist")
	}

	return pdFormat, nil
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
		return nil, newMissingArgumentError(node.name, FilterLevelsAttrId)
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
		return nil, newMissingArgumentError(node.name, FilePathId)
	}

	fileWriter, err := newFileWriter(path)
	if err != nil {
		return nil, err
	}

	return newFormattedWriter(fileWriter, currentFormat)
}

// Creates new SMTP writer if encountered in the config file.
func createSmtpWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
	err := checkUnexpectedAttribute(node, OutputFormatId, SenderAddressId, SenderNameId, HostNameId, HostPortId, UserNameId, UserPassId)
	if err != nil {
		return nil, err
	}
	// Node must have children.
	if !node.hasChildren() {
		return nil, nodeMustHaveChildrenError
	}
	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}
	senderAddress, ok := node.attributes[SenderAddressId]
	if !ok {
		return nil, newMissingArgumentError(node.name, SenderAddressId)
	}
	senderName, ok := node.attributes[SenderNameId]
	if !ok {
		return nil, newMissingArgumentError(node.name, SenderNameId)
	}
	// Process child nodes scanning for recipient email addresses and/or CA certificate paths.
	var recipientAddresses []string
	var caCertificatePaths []string
	for _, childNode := range node.children {
		switch childNode.name {
		// Extract recipient address from child nodes.
		case RecipientId:
			address, ok := childNode.attributes[AddressId]
			if !ok {
				return nil, newMissingArgumentError(childNode.name, AddressId)
			}
			recipientAddresses = append(recipientAddresses, address)
		// Extract CA certificate file path from child nodes.
		case CACertificatePathsId:
			path, ok := childNode.attributes[FilePathId]
			if !ok {
				return nil, newMissingArgumentError(childNode.name, FilePathId)
			}
			caCertificatePaths = append(caCertificatePaths, path)
		default:
			return nil, newUnexpectedChildElementError(childNode.name)
		}
	}
	hostName, ok := node.attributes[HostNameId]
	if !ok {
		return nil, newMissingArgumentError(node.name, HostNameId)
	}
	hostPort, ok := node.attributes[HostPortId]
	if !ok {
		return nil, newMissingArgumentError(node.name, HostPortId)
	}
	// Check if the string can really be converted into int.
	if _, err := strconv.Atoi(hostPort); err != nil {
		return nil, errors.New("Invalid host port number")
	}
	userName, ok := node.attributes[UserNameId]
	if !ok {
		return nil, newMissingArgumentError(node.name, UserNameId)
	}
	userPass, ok := node.attributes[UserPassId]
	if !ok {
		return nil, newMissingArgumentError(node.name, UserPassId)
	}
	smtpWriter, err := newSmtpWriter(
		senderAddress,
		senderName,
		recipientAddresses,
		hostName,
		hostPort,
		userName,
		userPass,
		caCertificatePaths,
	)
	if err != nil {
		return nil, err
	}
	return newFormattedWriter(smtpWriter, currentFormat)
}

func createConsoleWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
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

func createconnWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
	if node.hasChildren() {
		return nil, nodeCannotHaveChildrenError
	}

	err := checkUnexpectedAttribute(node, OutputFormatId, ConnWriterAddrAttr, ConnWriterNetAttr, ConnWriterReconnectOnMsgAttr)
	if err != nil {
		return nil, err
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	addr, isAddr := node.attributes[ConnWriterAddrAttr]
	if !isAddr {
		return nil, newMissingArgumentError(node.name, ConnWriterAddrAttr)
	}

	net, isNet := node.attributes[ConnWriterNetAttr]
	if !isNet {
		return nil, newMissingArgumentError(node.name, ConnWriterNetAttr)
	}

	reconnectOnMsg := false
	reconnectOnMsgStr, isReconnectOnMsgStr := node.attributes[ConnWriterReconnectOnMsgAttr]
	if isReconnectOnMsgStr {
		if reconnectOnMsgStr == "true" {
			reconnectOnMsg = true
		} else if reconnectOnMsgStr == "false" {
			reconnectOnMsg = false
		} else {
			return nil, errors.New("Node '" + node.name + "' has incorrect '" + ConnWriterReconnectOnMsgAttr + "' attribute value")
		}
	}

	connWriter := newConnWriter(net, addr, reconnectOnMsg)

	return newFormattedWriter(connWriter, currentFormat)
}

func createRollingFileWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (interface{}, error) {
	if node.hasChildren() {
		return nil, nodeCannotHaveChildrenError
	}

	rollingTypeStr, isRollingType := node.attributes[RollingFileTypeAttr]
	if !isRollingType {
		return nil, newMissingArgumentError(node.name, RollingFileTypeAttr)
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
		return nil, newMissingArgumentError(node.name, RollingFilePathAttr)
	}

	rollingArchiveStr, archiveAttrExists := node.attributes[RollingFileArchiveAttr]

	var rArchiveType rollingArchiveTypes 
	var rArchivePath string
	if !archiveAttrExists {
		rArchiveType = RollingArchiveNone
		rArchivePath = ""
	} else {
		rArchiveType, ok = rollingArchiveTypeFromString(rollingArchiveStr)
		if !ok {
			return nil, errors.New("Unknown rolling archive type: " + rollingArchiveStr)
		}

		if rArchiveType == RollingArchiveNone {
			rArchivePath = ""
		} else {
			rArchivePath, ok = node.attributes[RollingFileArchivePathAttr]
			if !ok {
				rArchivePath, ok = rollingArchiveTypesDefaultNames[rArchiveType]
				if !ok {
					return nil, fmt.Errorf("Cannot get default filename for archive type = %v", 
										   rArchiveType)
				}
			}
		}
	}

	if rollingType == RollingTypeSize {
		err := checkUnexpectedAttribute(node, OutputFormatId, RollingFileTypeAttr, RollingFilePathAttr,
									    RollingFileMaxSizeAttr, RollingFileMaxRollsAttr, RollingFileArchiveAttr,
									    RollingFileArchivePathAttr)
		if err != nil {
			return nil, err
		}

		maxSizeStr, isMaxSize := node.attributes[RollingFileMaxSizeAttr]
		if !isMaxSize {
			return nil, newMissingArgumentError(node.name, RollingFileMaxSizeAttr)
		}

		maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
		if err != nil {
			return nil, err
		}

		maxRollsStr, isMaxRolls := node.attributes[RollingFileMaxRollsAttr]
		if !isMaxRolls {
			return nil, newMissingArgumentError(node.name, RollingFileMaxRollsAttr)
		}

		maxRolls, err := strconv.Atoi(maxRollsStr)
		if err != nil {
			return nil, err
		}

		rollingWriter, err := newRollingFileWriterSize(path, rArchiveType, rArchivePath, maxSize, maxRolls)
		if err != nil {
			return nil, err
		}

		return newFormattedWriter(rollingWriter, currentFormat)

	} else if rollingType == RollingTypeDate {
		err := checkUnexpectedAttribute(node, OutputFormatId, RollingFileTypeAttr, RollingFilePathAttr, 
										RollingFileDataPatternAttr, RollingFileArchiveAttr,
										RollingFileArchivePathAttr)
		if err != nil {
			return nil, err
		}

		dataPattern, isDataPattern := node.attributes[RollingFileDataPatternAttr]
		if !isDataPattern {
			return nil, newMissingArgumentError(node.name, RollingFileDataPatternAttr)
		}

		rollingWriter, err := newRollingFileWriterDate(path, rArchiveType, rArchivePath, dataPattern)
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
		return nil, newMissingArgumentError(node.name, BufferedSizeAttr)
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
		return nil, errors.New("Inner writer cannot have his own format")
	}

	bufferedWriter, err := newBufferedWriter(formattedWriter.Writer(), size, time.Duration(flushPeriod))
	if err != nil {
		return nil, err
	}

	return newFormattedWriter(bufferedWriter, currentFormat)
}

// Returns an error if node has any attributes not listed in expectedAttrs.
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
			return newUnexpectedAttributeError(node.name, attr)
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
