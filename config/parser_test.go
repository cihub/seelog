// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"testing"
	//"reflect"
	"github.com/cihub/seelog/dispatchers"
	"github.com/cihub/seelog/writers"
	. "github.com/cihub/seelog/common"
	"github.com/cihub/seelog/test"
	"github.com/cihub/seelog/format"
	"strings"
	"fmt"
)

var parserTests []parserTest

type parserTest struct {
	testName      string
	config        string
	expected      interface{}
	errorExpected bool
}

func getParserTests() []parserTest {
	if parserTests == nil {
		parserTests = make([]parserTest, 0)

		testName := "Simple file output"
		testConfig := `
<seelog>
	<outputs>
		<file path="log.log"/>
	</outputs>
</seelog>`
		testExpected := new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testFileWriter, _ := writers.NewFileWriter("log.log")
		testHeadSplitter, _ := dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testFileWriter})
		testExpected.LogType = AsyncLoopLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Filter dispatcher"
		testConfig = `
<seelog type="sync">
	<outputs>
		<filter levels="debug, info, critical">
			<file path="log.log"/>
		</filter>
	</outputs>
</seelog>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testFileWriter, _ = writers.NewFileWriter("log.log")
		testFilter, _ := dispatchers.NewFilterDispatcher(format.DefaultFormatter, []interface{}{testFileWriter}, DebugLvl, InfoLvl, CriticalLvl)
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testFilter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Console writer"
		testConfig = `
<seelog type="sync">
	<outputs>
		<console />
	</outputs>
</seelog>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ := writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Default output"
		testConfig = `
<seelog type="sync"/>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Asyncloop behavior"
		testConfig = `
<seelog type="asyncloop"/>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = AsyncLoopLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Asynctimer behavior"
		testConfig = `
<seelog type="asynctimer" asyncinterval="101"/>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = AsyncTimerLoggerType
		testExpected.LoggerData = AsyncTimerLoggerData{101}
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Rolling file writer size"
		testConfig = `
<seelog type="sync">
	<outputs>
		<rollingfile type="size" filename="log.log" maxsize="100" maxrolls="5" />
	</outputs>
</seelog>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testRollingFileWriter, _ := writers.NewRollingFileWriterSize("log.log", 100, 5)
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testRollingFileWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Rolling file writer date"
		testConfig = `
<seelog type="sync">
	<outputs>
		<rollingfile type="date" filename="log.log" datepattern="2006-01-02T15:04:05Z07:00" />
	</outputs>
</seelog>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testRollingFileWriter, _ = writers.NewRollingFileWriterDate("log.log", "2006-01-02T15:04:05Z07:00")
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testRollingFileWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Buffered writer"
		testConfig = `
<seelog type="sync">
	<outputs>
		<buffered size="100500" flushperiod="100">
			<rollingfile type="date" filename="log.log" datepattern="2006-01-02T15:04:05Z07:00" />
		</buffered>
	</outputs>
</seelog>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testRollingFileWriter, _ = writers.NewRollingFileWriterDate("log.log", "2006-01-02T15:04:05Z07:00")
		testBufferedWriter, _ := writers.NewBufferedWriter(testRollingFileWriter, 100500, 100)
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testBufferedWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Inner splitter output"
		testConfig = `
<seelog type="sync">
	<outputs>
		<file path="log.log"/>
		<splitter>
			<file path="log1.log"/>
			<file path="log2.log"/>
		</splitter>
	</outputs>
</seelog>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testFileWriter1, _ := writers.NewFileWriter("log1.log")
		testFileWriter2, _ := writers.NewFileWriter("log2.log")
		testInnerSplitter, _ := dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testFileWriter1, testFileWriter2})
		testFileWriter, _ = writers.NewFileWriter("log.log")
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testFileWriter, testInnerSplitter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Format"
		testConfig = `
<seelog type="sync">
	<outputs formatid="dateFormat">
		<file path="log.log"/>
	</outputs>
	<formats>
		<format id="dateFormat" format="%Level %Msg %File" />
	</formats>
</seelog>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testFileWriter, _ = writers.NewFileWriter("log.log")
		testFormat, _ := format.NewFormatter("%Level %Msg %File")
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(testFormat, []interface{}{testFileWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Format2"
		testConfig = `
<seelog type="sync">
	<outputs formatid="format1">
		<file path="log.log"/>
		<file formatid="format2" path="log1.log"/>
	</outputs>
	<formats>
		<format id="format1" format="%Level %Msg %File" />
		<format id="format2" format="%l %Msg" />
	</formats>
</seelog>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testFileWriter, _ = writers.NewFileWriter("log.log")
		testFileWriter1, _ = writers.NewFileWriter("log1.log")
		testFormat1, _ := format.NewFormatter("%Level %Msg %File")
		testFormat2, _ := format.NewFormatter("%l %Msg")
		formattedWriter, _ := dispatchers.NewFormattedWriter(testFileWriter1, testFormat2)
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(testFormat1, []interface{}{testFileWriter, formattedWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Minlevel = warn"
		testConfig = `<seelog minlevel="warn"/>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(WarnLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = AsyncLoopLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Maxlevel = trace"
		testConfig = `<seelog maxlevel="trace"/>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, TraceLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = AsyncLoopLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Level between info and error"
		testConfig = `<seelog minlevel="info" maxlevel="error"/>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(InfoLvl, ErrorLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = AsyncLoopLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Off with minlevel"
		testConfig = `<seelog minlevel="off"/>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewOffConstraints()
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = AsyncLoopLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Off with levels"
		testConfig = `<seelog levels="off"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Levels list"
		testConfig = `<seelog levels="debug, info, critical"/>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewListConstraints([]LogLevel{
			DebugLvl, InfoLvl, CriticalLvl})
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = AsyncLoopLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Errors #1"
		testConfig = `<seelog minlevel="debug" minlevel="trace"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #2"
		testConfig = `<seelog minlevel="error" maxlevel="debug"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #3"
		testConfig = `<seelog maxlevel="debug" maxlevel="trace"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #4"
		testConfig = `<seelog maxlevel="off"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #5"
		testConfig = `<seelog minlevel="off" maxlevel="trace"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #6"
		testConfig = `<seelog minlevel="warn" maxlevel="error" levels="debug"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #7"
		testConfig = `<not_seelog/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #8"
		testConfig = `<seelog levels="warn, debug, test"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #9"
		testConfig = `<seelog levels=""/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #10"
		testConfig = `<seelog levels="off" something="abc"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #11"
		testConfig = `<seelog><output/></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #12"
		testConfig = `<seelog><outputs/><outputs/></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #13"
		testConfig = `<seelog><exceptions/></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #14"
		testConfig = `<seelog><formats/></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #15"
		testConfig = `<seelog><outputs><splitter/></outputs></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #16"
		testConfig = `<seelog><outputs><filter/></outputs></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #17"
		testConfig = `<seelog><outputs><file path="log.log"><something/></file></outputs></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #18"
		testConfig = `<seelog><outputs><buffered size="100500" flushperiod="100"/></outputs></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #19"
		testConfig = `<seelog><outputs></outputs></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Exceptions: restricting"
		testConfig =
			`
<seelog type="sync">
	<exceptions>
		<exception funcpattern="Test*" filepattern="someFile.go" minlevel="off"/>
	</exceptions>
</seelog>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		listConstraint, _ := NewOffConstraints()
		exception, _ := NewLogLevelException("Test*", "someFile.go", listConstraint)
		testExpected.Exceptions = []*LogLevelException{exception}
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Exceptions: allowing #1"
		testConfig =
			`
<seelog type="sync" levels="error">
	<exceptions>
		<exception filepattern="testfile.go" minlevel="trace"/>
	</exceptions>
</seelog>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewListConstraints([]LogLevel{ErrorLvl})
		minMaxConstraint, _ := NewMinMaxConstraints(TraceLvl, CriticalLvl)
		exception, _ = NewLogLevelException("*", "testfile.go", minMaxConstraint)
		testExpected.Exceptions = []*LogLevelException{exception}
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Exceptions: allowing #2"
		testConfig =
			`
<seelog type="sync" levels="off">
	<exceptions>
		<exception filepattern="testfile.go" minlevel="warn"/>
	</exceptions>
</seelog>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewOffConstraints()
		minMaxConstraint, _ = NewMinMaxConstraints(WarnLvl, CriticalLvl)
		exception, _ = NewLogLevelException("*", "testfile.go", minMaxConstraint)
		testExpected.Exceptions = []*LogLevelException{exception}
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{testConsoleWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Errors #11"
		testConfig = `
<seelog type="sync"><exceptions>
		<exception filepattern="testfile.go" minlevel="trace"/>
		<exception filepattern="testfile.go" minlevel="warn"/>
</exceptions></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #12"
		testConfig = `
<seelog type="sync"><exceptions>
		<exception filepattern="!@+$)!!%&@(^$" minlevel="trace"/>
</exceptions></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #13"
		testConfig = `
<seelog type="sync"><exceptions>
		<exception filepattern="*" minlevel="unknown"/>
</exceptions></seelog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #14"
		testConfig = `
<seelog type="sync" levels=”off”>
	<exceptions>
		<exception filepattern="testfile.go" minlevel="off"/>
	</exceptions>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #15"
		testConfig = `
<seelog type="sync" levels=”trace”>
	<exceptions>
		<exception filepattern="testfile.go" levels="trace"/>
	</exceptions>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #16"
		testConfig = `
<seelog type="sync" minlevel=”trace”>
	<exceptions>
		<exception filepattern="testfile.go" minlevel="trace"/>
	</exceptions>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #17"
		testConfig = `
<seelog type="sync" minlevel=”trace”>
	<exceptions>
		<exception filepattern="testfile.go" minlevel="warn"/>
	</exceptions>
	<exceptions>
		<exception filepattern="testfile.go" minlevel="warn"/>
	</exceptions>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #18"
		testConfig = `
<seelog type="sync" minlevel=”trace”>
	<exceptions>
		<exception filepattern="testfile.go"/>
	</exceptions>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #19"
		testConfig = `
<seelog type="sync" minlevel=”trace”>
	<exceptions>
		<exception minlevel="warn"/>
	</exceptions>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #20"
		testConfig = `
<seelog type="sync" minlevel=”trace”>
	<exceptions>
		<exception/>
	</exceptions>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #21"
		testConfig = `
<seelog>
	<outputs>
		<splitter>
		</splitter>
	</outputs>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #22"
		testConfig = `
<seelog type="sync">
	<outputs>
		<filter levels="debug, info, critical">

		</filter>
	</outputs>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #23"
		testConfig = `
<seelog type="sync">
	<outputs>
		<buffered size="100500" flushperiod="100">

		</buffered>
	</outputs>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #24"
		testConfig = `
<seelog type="sync">
	<outputs>
		<buffered size="100500" flushperiod="100">
			<rollingfile type="date" filename="log.log" datepattern="2006-01-02T15:04:05Z07:00" formatid="testFormat"/>
		</buffered>
	</outputs>
	<formats>
		<format id="testFormat" format="%Level %Msg %File 123" />
	</formats>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Errors #25"
		testConfig = `
<seelog type="sync">
	<outputs>
		<outputs>
			<file path="file.log"/>
		</outputs>
		<outputs>
			<file path="file.log"/>
		</outputs>
	</outputs>
</seelog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
		
		testName = "Buffered writer same formatid override"
		testConfig = `
<seelog type="sync">
	<outputs>
		<buffered size="100500" flushperiod="100" formatid="testFormat">
			<rollingfile type="date" filename="log.log" datepattern="2006-01-02T15:04:05Z07:00" formatid="testFormat"/>
		</buffered>
	</outputs>
	<formats>
		<format id="testFormat" format="%Level %Msg %File 123" />
	</formats>
</seelog>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testRollingFileWriter, _ = writers.NewRollingFileWriterDate("log.log", "2006-01-02T15:04:05Z07:00")
		testBufferedWriter, _ = writers.NewBufferedWriter(testRollingFileWriter, 100500, 100)
		testFormat, _ = format.NewFormatter("%Level %Msg %File 123")
		formattedWriter, _ = dispatchers.NewFormattedWriter(testBufferedWriter, testFormat)
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher(format.DefaultFormatter, []interface{}{formattedWriter})
		testExpected.LogType = SyncLoggerType
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})
		

	}

	return parserTests
}

// We are waiting for structs equality (Planned in Go 1 release) and this func is a
// temporary solution
func configsAreEqual(conf1 *LogConfig, conf2 interface{}) bool {
	if conf1 == nil {
		return conf2 == nil
	}
	if conf2 == nil {
		return conf1 == nil
	}
	logConfig, ok := conf2.(*LogConfig)
	
	if !ok {
		return false
	}
	
	return fmt.Sprintf("%s", conf1) == fmt.Sprintf("%s", logConfig)
}

func TestParser(t *testing.T) {

	testFSWrapper, err := test.NewEmptyFSTestWrapper()

	if err != nil {
		t.Fatalf("Fatal error in test fs initialization: %s", err.Error())
	}

	writers.SetTestMode(testFSWrapper)

	for _, test := range getParserTests() {

		conf, err := ConfigFromReader(strings.NewReader(test.config))

		if (err != nil) != test.errorExpected {
			t.Errorf("\n----ERROR in %s:\nConfig: %s\n* Expected error:%t. Got error: %t\n", test.testName,
				test.config, test.errorExpected, (err != nil))
			if err != nil {
				t.Logf("%s\n", err.Error())
			}
			continue
		}

		if err == nil && !configsAreEqual(conf, test.expected) {
			t.Errorf("\n----ERROR in %s:\nConfig: %s\n* Expected: %s. \n* Got: %s\n", test.testName, test.config, test.expected, conf)
		}
	}
}
