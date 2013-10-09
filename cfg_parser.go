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

// Names of elements of seelog config.
const (
	seelogConfigId                  = "seelog"
	outputsId                       = "outputs"
	formatsId                       = "formats"
	minLevelId                      = "minlevel"
	maxLevelId                      = "maxlevel"
	levelsId                        = "levels"
	exceptionsId                    = "exceptions"
	exceptionId                     = "exception"
	funcPatternId                   = "funcpattern"
	filePatternId                   = "filepattern"
	formatId                        = "format"
	formatAttrId                    = "format"
	formatKeyAttrId                 = "id"
	outputFormatId                  = "formatid"
	pathId                          = "path"
	fileWriterId                    = "file"
	smtpWriterId                    = "smtp"
	senderaddressId                 = "senderaddress"
	senderNameId                    = "sendername"
	recipientId                     = "recipient"
	addressId                       = "address"
	hostNameId                      = "hostname"
	hostPortId                      = "hostport"
	userNameId                      = "username"
	userPassId                      = "password"
	cACertDirpathId                 = "cacertdirpath"
	splitterDispatcherId            = "splitter"
	consoleWriterId                 = "console"
	customReceiverId                = "custom"
	customNameAttrId                = "name"
	customNameDataAttrPrefix        = "data-"
	filterDispatcherId              = "filter"
	filterLevelsAttrId              = "levels"
	rollingfileWriterId             = "rollingfile"
	rollingFileTypeAttr             = "type"
	rollingFilePathAttr             = "filename"
	rollingFileMaxSizeAttr          = "maxsize"
	rollingFileMaxRollsAttr         = "maxrolls"
	rollingFileDataPatternAttr      = "datepattern"
	rollingFileArchiveAttr          = "archivetype"
	rollingFileArchivePathAttr      = "archivepath"
	bufferedWriterId                = "buffered"
	bufferedSizeAttr                = "size"
	bufferedFlushPeriodAttr         = "flushperiod"
	loggerTypeFromStringAttr        = "type"
	asyncLoggerIntervalAttr         = "asyncinterval"
	adaptLoggerMinIntervalAttr      = "mininterval"
	adaptLoggerMaxIntervalAttr      = "maxinterval"
	adaptLoggerCriticalMsgCountAttr = "critmsgcount"
	predefinedPrefix                = "std:"
	connWriterId                    = "conn"
	connWriterAddrAttr              = "addr"
	connWriterNetAttr               = "net"
	connWriterReconnectOnMsgAttr    = "reconnectonmsg"
)

// CustomReceiverProducer is the signature of the function CfgParseParams needs to create
// custom receivers.
type CustomReceiverProducer func(CustomReceiverInitArgs) (CustomReceiver, error)

// CfgParseParams represent specific parse options or flags used by parser. It is used if seelog parser needs
// some special directives or additional info to correctly parse a config.
type CfgParseParams struct {
	// CustomReceiverProducers expose the same functionality as RegisterReceiver func
	// but only in the scope (context) of the config parse func instead of a global package scope.
	//
	// It means that if you use custom receivers in your code, you may either register them globally once with
	// RegisterReceiver or you may call funcs like LoggerFromParamConfigAsFile (with 'ParamConfig')
	// and use CustomReceiverProducers to provide custom producer funcs.
	//
	// A producer func is called when config parser processes a '<custom>' element. It takes the 'name' attribute
	// of the element and tries to find a match in two places:
	// 1) CfgParseParams.CustomReceiverProducers map
	// 2) Global type map, filled by RegisterReceiver
	//
	// If a match is found in the CustomReceiverProducers map, parser calls the corresponding producer func
	// passing the init args to it.	The func takes exactly the same args as CustomReceiver.AfterParse.
	// The producer func must return a correct receiver or an error. If case of error, seelog will behave
	// in the same way as with any other config error.
	//
	// You may use this param to set custom producers in case you need to pass some context when instantiating
	// a custom receiver or if you frequently change custom receivers with different parameters or in any other
	// situation where package-level registering (RegisterReceiver) is not an option for you.
	CustomReceiverProducers map[string]CustomReceiverProducer
}

func (cfg *CfgParseParams) String() string {
	return fmt.Sprintf("CfgParams: {custom_recs=%d}", len(cfg.CustomReceiverProducers))
}

type elementMapEntry struct {
	constructor func(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter, cfg *CfgParseParams) (interface{}, error)
}

var elementMap map[string]elementMapEntry
var predefinedFormats map[string]*formatter

func init() {
	elementMap = map[string]elementMapEntry{
		fileWriterId:         {createfileWriter},
		splitterDispatcherId: {createSplitter},
		customReceiverId:     {createCustomReceiver},
		filterDispatcherId:   {createFilter},
		consoleWriterId:      {createConsoleWriter},
		rollingfileWriterId:  {createRollingFileWriter},
		bufferedWriterId:     {createbufferedWriter},
		smtpWriterId:         {createSmtpWriter},
		connWriterId:         {createconnWriter},
	}

	err := fillPredefinedFormats()
	if err != nil {
		panic(fmt.Sprintf("Seelog couldn't start: predefined formats creation failed. Error: %s", err.Error()))
	}
}

func fillPredefinedFormats() error {
	predefinedFormatsWithoutPrefix := map[string]string{
		"xml-debug":       `<time>%Ns</time><lev>%Lev</lev><msg>%Msg</msg><path>%RelFile</path><func>%Func</func><line>%Line</line>`,
		"xml-debug-short": `<t>%Ns</t><l>%l</l><m>%Msg</m><p>%RelFile</p><f>%Func</f>`,
		"xml":             `<time>%Ns</time><lev>%Lev</lev><msg>%Msg</msg>`,
		"xml-short":       `<t>%Ns</t><l>%l</l><m>%Msg</m>`,

		"json-debug":       `{"time":%Ns,"lev":"%Lev","msg":"%Msg","path":"%RelFile","func":"%Func","line":"%Line"}`,
		"json-debug-short": `{"t":%Ns,"l":"%Lev","m":"%Msg","p":"%RelFile","f":"%Func"}`,
		"json":             `{"time":%Ns,"lev":"%Lev","msg":"%Msg"}`,
		"json-short":       `{"t":%Ns,"l":"%Lev","m":"%Msg"}`,

		"debug":       `[%LEVEL] %RelFile:%Func.%Line %Date %Time %Msg%n`,
		"debug-short": `[%LEVEL] %Date %Time %Msg%n`,
		"fast":        `%Ns %l %Msg%n`,
	}

	predefinedFormats = make(map[string]*formatter)

	for formatKey, format := range predefinedFormatsWithoutPrefix {
		formatter, err := newFormatter(format)
		if err != nil {
			return err
		}

		predefinedFormats[predefinedPrefix+formatKey] = formatter
	}

	return nil
}

// configFromReader parses data from a given reader.
// Returns parsed config which can be used to create logger in case no errors occured.
// Returns error if format is incorrect or anything happened.
func configFromReader(reader io.Reader) (*logConfig, error) {
	return configFromReaderWithConfig(reader, nil)
}

// configFromReader parses data from a given reader.
// Returns parsed config which can be used to create logger in case no errors occured.
// Returns error if format is incorrect or anything happened.
func configFromReaderWithConfig(reader io.Reader, cfg *CfgParseParams) (*logConfig, error) {
	config, err := unmarshalConfig(reader)
	if err != nil {
		return nil, err
	}

	if config.name != seelogConfigId {
		return nil, errors.New("Root xml tag must be '" + seelogConfigId + "'")
	}

	err = checkUnexpectedAttribute(
		config,
		minLevelId,
		maxLevelId,
		levelsId,
		loggerTypeFromStringAttr,
		asyncLoggerIntervalAttr,
		adaptLoggerMinIntervalAttr,
		adaptLoggerMaxIntervalAttr,
		adaptLoggerCriticalMsgCountAttr,
	)
	if err != nil {
		return nil, err
	}

	err = checkExpectedElements(config, optionalElement(outputsId), optionalElement(formatsId), optionalElement(exceptionsId))
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

	dispatcher, err := getOutputsTree(config, formats, cfg)
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

	return newConfig(constraints, exceptions, dispatcher, loggerType, logData, cfg)
}

func getConstraints(node *xmlNode) (logLevelConstraints, error) {
	minLevelStr, isMinLevel := node.attributes[minLevelId]
	maxLevelStr, isMaxLevel := node.attributes[maxLevelId]
	levelsStr, isLevels := node.attributes[levelsId]

	if isLevels && (isMinLevel && isMaxLevel) {
		return nil, errors.New("For level declaration use '" + levelsId + "'' OR '" + minLevelId +
			"', '" + maxLevelId + "'")
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
			return nil, errors.New("Declared " + minLevelId + " not found: " + minLevelStr)
		}
	}

	var maxLevel LogLevel = CriticalLvl
	if isMaxLevel {
		found := true
		maxLevel, found = LogLevelFromString(maxLevelStr)
		if !found {
			return nil, errors.New("Declared " + maxLevelId + " not found: " + maxLevelStr)
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
		if child.name == exceptionsId {
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
		if exceptionNode.name != exceptionId {
			return nil, errors.New("Incorrect nested element in exceptions section: " + exceptionNode.name)
		}

		err := checkUnexpectedAttribute(exceptionNode, minLevelId, maxLevelId, levelsId, funcPatternId, filePatternId)
		if err != nil {
			return nil, err
		}

		constraints, err := getConstraints(exceptionNode)
		if err != nil {
			return nil, errors.New("Incorrect " + exceptionsId + " node: " + err.Error())
		}

		funcPattern, isFuncPattern := exceptionNode.attributes[funcPatternId]
		filePattern, isFilePattern := exceptionNode.attributes[filePatternId]
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
		if child.name == formatsId {
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
		if formatNode.name != formatId {
			return nil, errors.New("Incorrect nested element in " + formatsId + " section: " + formatNode.name)
		}

		err := checkUnexpectedAttribute(formatNode, formatKeyAttrId, formatId)
		if err != nil {
			return nil, err
		}

		id, isId := formatNode.attributes[formatKeyAttrId]
		formatStr, isFormat := formatNode.attributes[formatAttrId]
		if !isId {
			return nil, errors.New("Format has no '" + formatKeyAttrId + "' attribute")
		}
		if !isFormat {
			return nil, errors.New("Format[" + id + "] has no '" + formatAttrId + "' attribute")
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
	logTypeStr, loggerTypeExists := config.attributes[loggerTypeFromStringAttr]

	if !loggerTypeExists {
		return defaultloggerTypeFromString, nil, nil
	}

	logType, found := getLoggerTypeFromString(logTypeStr)

	if !found {
		return 0, nil, errors.New(fmt.Sprintf("Unknown logger type: %s", logTypeStr))
	}

	if logType == asyncTimerloggerTypeFromString {
		intervalStr, intervalExists := config.attributes[asyncLoggerIntervalAttr]
		if !intervalExists {
			return 0, nil, newMissingArgumentError(config.name, asyncLoggerIntervalAttr)
		}

		interval, err := strconv.ParseUint(intervalStr, 10, 32)
		if err != nil {
			return 0, nil, err
		}

		logData = asyncTimerLoggerData{uint32(interval)}
	} else if logType == adaptiveLoggerTypeFromString {

		// Min interval
		minIntStr, minIntExists := config.attributes[adaptLoggerMinIntervalAttr]
		if !minIntExists {
			return 0, nil, newMissingArgumentError(config.name, adaptLoggerMinIntervalAttr)
		}
		minInterval, err := strconv.ParseUint(minIntStr, 10, 32)
		if err != nil {
			return 0, nil, err
		}

		// Max interval
		maxIntStr, maxIntExists := config.attributes[adaptLoggerMaxIntervalAttr]
		if !maxIntExists {
			return 0, nil, newMissingArgumentError(config.name, adaptLoggerMaxIntervalAttr)
		}
		maxInterval, err := strconv.ParseUint(maxIntStr, 10, 32)
		if err != nil {
			return 0, nil, err
		}

		// Critical msg count
		criticalMsgCountStr, criticalMsgCountExists := config.attributes[adaptLoggerCriticalMsgCountAttr]
		if !criticalMsgCountExists {
			return 0, nil, newMissingArgumentError(config.name, adaptLoggerCriticalMsgCountAttr)
		}
		criticalMsgCount, err := strconv.ParseUint(criticalMsgCountStr, 10, 32)
		if err != nil {
			return 0, nil, err
		}

		logData = adaptiveLoggerData{uint32(minInterval), uint32(maxInterval), uint32(criticalMsgCount)}
	}

	return logType, logData, nil
}

func getOutputsTree(config *xmlNode, formats map[string]*formatter, cfg *CfgParseParams) (dispatcherInterface, error) {
	var outputsNode *xmlNode
	for _, child := range config.children {
		if child.name == outputsId {
			outputsNode = child
			break
		}
	}

	if outputsNode != nil {
		err := checkUnexpectedAttribute(outputsNode, outputFormatId)
		if err != nil {
			return nil, err
		}

		formatter, err := getCurrentFormat(outputsNode, defaultformatter, formats)
		if err != nil {
			return nil, err
		}

		output, err := createSplitter(outputsNode, formatter, formats, cfg)
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
	return newSplitDispatcher(defaultformatter, []interface{}{console})
}

func getCurrentFormat(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter) (*formatter, error) {
	formatId, isFormatId := node.attributes[outputFormatId]
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

func createInnerReceivers(node *xmlNode, format *formatter, formats map[string]*formatter, cfg *CfgParseParams) ([]interface{}, error) {
	outputs := make([]interface{}, 0)
	for _, childNode := range node.children {
		entry, ok := elementMap[childNode.name]
		if !ok {
			return nil, errors.New("Unnknown tag '" + childNode.name + "' in outputs section")
		}

		output, err := entry.constructor(childNode, format, formats, cfg)
		if err != nil {
			return nil, err
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

func createSplitter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter, cfg *CfgParseParams) (interface{}, error) {
	err := checkUnexpectedAttribute(node, outputFormatId)
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

	receivers, err := createInnerReceivers(node, currentFormat, formats, cfg)
	if err != nil {
		return nil, err
	}

	return newSplitDispatcher(currentFormat, receivers)
}

func createCustomReceiver(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter, cfg *CfgParseParams) (interface{}, error) {
	dataCustomPrefixes := make(map[string]string)
	// Expecting only 'formatid', 'name' and 'data-' attrs
	for attr, attrval := range node.attributes {
		isExpected := false
		if attr == outputFormatId ||
			attr == customNameAttrId {
			isExpected = true
		}
		if strings.HasPrefix(attr, customNameDataAttrPrefix) {
			dataCustomPrefixes[attr[len(customNameDataAttrPrefix):]] = attrval
			isExpected = true
		}
		if !isExpected {
			return nil, newUnexpectedAttributeError(node.name, attr)
		}
	}

	if node.hasChildren() {
		return nil, nodeCannotHaveChildrenError
	}
	customName, hasCustomName := node.attributes[customNameAttrId]
	if !hasCustomName {
		return nil, newMissingArgumentError(node.name, customNameAttrId)
	}
	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}
	args := CustomReceiverInitArgs{
		XmlCustomAttrs: dataCustomPrefixes,
	}

	if cfg != nil && cfg.CustomReceiverProducers != nil {
		if prod, ok := cfg.CustomReceiverProducers[customName]; ok {
			rec, err := prod(args)
			if err != nil {
				return nil, err
			}
			creceiver, err := newCustomReceiverDispatcherByValue(currentFormat, rec, customName, args)
			if err != nil {
				return nil, err
			}
			err = rec.AfterParse(args)
			if err != nil {
				return nil, err
			}
			return creceiver, nil
		}
	}

	return newCustomReceiverDispatcher(currentFormat, customName, args)
}

func createFilter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter, cfg *CfgParseParams) (interface{}, error) {
	err := checkUnexpectedAttribute(node, outputFormatId, filterLevelsAttrId)
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

	levelsStr, isLevels := node.attributes[filterLevelsAttrId]
	if !isLevels {
		return nil, newMissingArgumentError(node.name, filterLevelsAttrId)
	}

	levels, err := parseLevels(levelsStr)
	if err != nil {
		return nil, err
	}

	receivers, err := createInnerReceivers(node, currentFormat, formats, cfg)
	if err != nil {
		return nil, err
	}

	return newFilterDispatcher(currentFormat, receivers, levels...)
}

func createfileWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter, cfg *CfgParseParams) (interface{}, error) {
	err := checkUnexpectedAttribute(node, outputFormatId, pathId)
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

	path, isPath := node.attributes[pathId]
	if !isPath {
		return nil, newMissingArgumentError(node.name, pathId)
	}

	fileWriter, err := newFileWriter(path)
	if err != nil {
		return nil, err
	}

	return newFormattedWriter(fileWriter, currentFormat)
}

// Creates new SMTP writer if encountered in the config file.
func createSmtpWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter, cfg *CfgParseParams) (interface{}, error) {
	err := checkUnexpectedAttribute(node, outputFormatId, senderaddressId, senderNameId, hostNameId, hostPortId, userNameId, userPassId)
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
	senderAddress, ok := node.attributes[senderaddressId]
	if !ok {
		return nil, newMissingArgumentError(node.name, senderaddressId)
	}
	senderName, ok := node.attributes[senderNameId]
	if !ok {
		return nil, newMissingArgumentError(node.name, senderNameId)
	}
	// Process child nodes scanning for recipient email addresses and/or CA certificate paths.
	var recipientAddresses []string
	var caCertDirPaths []string
	for _, childNode := range node.children {
		switch childNode.name {
		// Extract recipient address from child nodes.
		case recipientId:
			address, ok := childNode.attributes[addressId]
			if !ok {
				return nil, newMissingArgumentError(childNode.name, addressId)
			}
			recipientAddresses = append(recipientAddresses, address)
		// Extract CA certificate file path from child nodes.
		case cACertDirpathId:
			path, ok := childNode.attributes[pathId]
			if !ok {
				return nil, newMissingArgumentError(childNode.name, pathId)
			}
			caCertDirPaths = append(caCertDirPaths, path)
		default:
			return nil, newUnexpectedChildElementError(childNode.name)
		}
	}
	hostName, ok := node.attributes[hostNameId]
	if !ok {
		return nil, newMissingArgumentError(node.name, hostNameId)
	}

	hostPort, ok := node.attributes[hostPortId]
	if !ok {
		return nil, newMissingArgumentError(node.name, hostPortId)
	}

	// Check if the string can really be converted into int.
	if _, err := strconv.Atoi(hostPort); err != nil {
		return nil, errors.New("Invalid host port number")
	}

	userName, ok := node.attributes[userNameId]
	if !ok {
		return nil, newMissingArgumentError(node.name, userNameId)
	}

	userPass, ok := node.attributes[userPassId]
	if !ok {
		return nil, newMissingArgumentError(node.name, userPassId)
	}

	smtpWriter := newSmtpWriter(
		senderAddress,
		senderName,
		recipientAddresses,
		hostName,
		hostPort,
		userName,
		userPass,
		caCertDirPaths,
	)

	return newFormattedWriter(smtpWriter, currentFormat)
}

func createConsoleWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter, cfg *CfgParseParams) (interface{}, error) {
	err := checkUnexpectedAttribute(node, outputFormatId)
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

func createconnWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter, cfg *CfgParseParams) (interface{}, error) {
	if node.hasChildren() {
		return nil, nodeCannotHaveChildrenError
	}

	err := checkUnexpectedAttribute(node, outputFormatId, connWriterAddrAttr, connWriterNetAttr, connWriterReconnectOnMsgAttr)
	if err != nil {
		return nil, err
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	addr, isAddr := node.attributes[connWriterAddrAttr]
	if !isAddr {
		return nil, newMissingArgumentError(node.name, connWriterAddrAttr)
	}

	net, isNet := node.attributes[connWriterNetAttr]
	if !isNet {
		return nil, newMissingArgumentError(node.name, connWriterNetAttr)
	}

	reconnectOnMsg := false
	reconnectOnMsgStr, isReconnectOnMsgStr := node.attributes[connWriterReconnectOnMsgAttr]
	if isReconnectOnMsgStr {
		if reconnectOnMsgStr == "true" {
			reconnectOnMsg = true
		} else if reconnectOnMsgStr == "false" {
			reconnectOnMsg = false
		} else {
			return nil, errors.New("Node '" + node.name + "' has incorrect '" + connWriterReconnectOnMsgAttr + "' attribute value")
		}
	}

	connWriter := newConnWriter(net, addr, reconnectOnMsg)

	return newFormattedWriter(connWriter, currentFormat)
}

func createRollingFileWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter, cfg *CfgParseParams) (interface{}, error) {
	if node.hasChildren() {
		return nil, nodeCannotHaveChildrenError
	}

	rollingTypeStr, isRollingType := node.attributes[rollingFileTypeAttr]
	if !isRollingType {
		return nil, newMissingArgumentError(node.name, rollingFileTypeAttr)
	}

	rollingType, ok := rollingTypeFromString(rollingTypeStr)
	if !ok {
		return nil, errors.New("Unknown rolling file type: " + rollingTypeStr)
	}

	currentFormat, err := getCurrentFormat(node, formatFromParent, formats)
	if err != nil {
		return nil, err
	}

	path, isPath := node.attributes[rollingFilePathAttr]
	if !isPath {
		return nil, newMissingArgumentError(node.name, rollingFilePathAttr)
	}

	rollingArchiveStr, archiveAttrExists := node.attributes[rollingFileArchiveAttr]

	var rArchiveType rollingArchiveType
	var rArchivePath string
	if !archiveAttrExists {
		rArchiveType = rollingArchiveNone
		rArchivePath = ""
	} else {
		rArchiveType, ok = rollingArchiveTypeFromString(rollingArchiveStr)
		if !ok {
			return nil, errors.New("Unknown rolling archive type: " + rollingArchiveStr)
		}

		if rArchiveType == rollingArchiveNone {
			rArchivePath = ""
		} else {
			rArchivePath, ok = node.attributes[rollingFileArchivePathAttr]
			if !ok {
				rArchivePath, ok = rollingArchiveTypesDefaultNames[rArchiveType]
				if !ok {
					return nil, fmt.Errorf("Cannot get default filename for archive type = %v",
						rArchiveType)
				}
			}
		}
	}

	if rollingType == rollingTypeSize {
		err := checkUnexpectedAttribute(node, outputFormatId, rollingFileTypeAttr, rollingFilePathAttr,
			rollingFileMaxSizeAttr, rollingFileMaxRollsAttr, rollingFileArchiveAttr,
			rollingFileArchivePathAttr)
		if err != nil {
			return nil, err
		}

		maxSizeStr, isMaxSize := node.attributes[rollingFileMaxSizeAttr]
		if !isMaxSize {
			return nil, newMissingArgumentError(node.name, rollingFileMaxSizeAttr)
		}

		maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
		if err != nil {
			return nil, err
		}

		maxRolls := 0
		maxRollsStr, isMaxRolls := node.attributes[rollingFileMaxRollsAttr]
		if isMaxRolls {
			maxRolls, err = strconv.Atoi(maxRollsStr)
			if err != nil {
				return nil, err
			}
		}

		rollingWriter, err := newRollingFileWriterSize(path, rArchiveType, rArchivePath, maxSize, maxRolls)
		if err != nil {
			return nil, err
		}

		return newFormattedWriter(rollingWriter, currentFormat)

	} else if rollingType == rollingTypeTime {
		err := checkUnexpectedAttribute(node, outputFormatId, rollingFileTypeAttr, rollingFilePathAttr,
			rollingFileDataPatternAttr, rollingFileArchiveAttr, rollingFileMaxRollsAttr,
			rollingFileArchivePathAttr)
		if err != nil {
			return nil, err
		}

		maxRolls := 0
		maxRollsStr, isMaxRolls := node.attributes[rollingFileMaxRollsAttr]
		if isMaxRolls {
			maxRolls, err = strconv.Atoi(maxRollsStr)
			if err != nil {
				return nil, err
			}
		}

		dataPattern, isDataPattern := node.attributes[rollingFileDataPatternAttr]
		if !isDataPattern {
			return nil, newMissingArgumentError(node.name, rollingFileDataPatternAttr)
		}

		rollingWriter, err := newRollingFileWriterTime(path, rArchiveType, rArchivePath, maxRolls, dataPattern, rollingIntervalAny)
		if err != nil {
			return nil, err
		}

		return newFormattedWriter(rollingWriter, currentFormat)
	}

	return nil, errors.New("Incorrect rolling writer type " + rollingTypeStr)
}

func createbufferedWriter(node *xmlNode, formatFromParent *formatter, formats map[string]*formatter, cfg *CfgParseParams) (interface{}, error) {
	err := checkUnexpectedAttribute(node, outputFormatId, bufferedSizeAttr, bufferedFlushPeriodAttr)
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

	sizeStr, isSize := node.attributes[bufferedSizeAttr]
	if !isSize {
		return nil, newMissingArgumentError(node.name, bufferedSizeAttr)
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return nil, err
	}

	flushPeriod := 0
	flushPeriodStr, isFlushPeriod := node.attributes[bufferedFlushPeriodAttr]
	if isFlushPeriod {
		flushPeriod, err = strconv.Atoi(flushPeriodStr)
		if err != nil {
			return nil, err
		}
	}

	// Inner writer couldn't have its own format, so we pass 'currentFormat' as its parent format
	receivers, err := createInnerReceivers(node, currentFormat, formats, cfg)
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
