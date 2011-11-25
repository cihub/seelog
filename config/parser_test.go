// Package config contains configuration functionality of sealog
package config

import (
	"testing"
	"reflect"
	"sealog/dispatchers"
	"sealog/writers"
	. "sealog/common"
	"sealog/test"
	"strings"
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
<!-- test 1 -->
<sealog>
	<outputs>
		<file path="log.log"/>
	</outputs>
</sealog>`
		testExpected := new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testFileWriter, _ := writers.NewFileWriter("log.log")
		testHeadSplitter, _ := dispatchers.NewSplitDispatcher([]interface{}{testFileWriter})
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Default output"
		testConfig = `
<sealog/>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ := writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher([]interface{}{testConsoleWriter})
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Minlevel = warn"
		testConfig = `<sealog minlevel="warn"/>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(WarnLvl, CriticalLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher([]interface{}{testConsoleWriter})
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Maxlevel = trace"
		testConfig = `<sealog maxlevel="trace"/>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, TraceLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher([]interface{}{testConsoleWriter})
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Level between info and error"
		testConfig = `<sealog minlevel="info" maxlevel="error"/>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(InfoLvl, ErrorLvl)
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher([]interface{}{testConsoleWriter})
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Off with minlevel"
		testConfig = `<sealog minlevel="off"/>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewOffConstraints()
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher([]interface{}{testConsoleWriter})
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Off with levels"
		testConfig = `<sealog levels="off"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Levels list"
		testConfig = `<sealog levels="debug, info, critical"/>`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewListConstraints([]LogLevel{
			DebugLvl, InfoLvl, CriticalLvl})
		testExpected.Exceptions = nil
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher([]interface{}{testConsoleWriter})
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Errors #1"
		testConfig = `<sealog minlevel="debug" minlevel="trace"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #2"
		testConfig = `<sealog minlevel="error" maxlevel="debug"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #3"
		testConfig = `<sealog maxlevel="debug" maxlevel="trace"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #4"
		testConfig = `<sealog maxlevel="off"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #5"
		testConfig = `<sealog minlevel="off" maxlevel="trace"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #6"
		testConfig = `<sealog minlevel="warn" maxlevel="error" levels="debug"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #7"
		testConfig = `<not_sealog/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #8"
		testConfig = `<sealog levels="warn, debug, test"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #9"
		testConfig = `<sealog levels=""/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #10"
		testConfig = `<sealog levels="off" something="abc"/>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Exceptions: restricting"
		testConfig =
			`
<sealog>
	<exceptions>
		<exception funcpattern="Test*" filepattern="someFile.go" minlevel="off"/>
	</exceptions>
</sealog>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewMinMaxConstraints(TraceLvl, CriticalLvl)
		listConstraint, _ := NewOffConstraints()
		exception, _ := NewLogLevelException("Test*", "someFile.go", listConstraint)
		testExpected.Exceptions = []*LogLevelException{exception}
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher([]interface{}{testConsoleWriter})
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Exceptions: allowing #1"
		testConfig =
			`
<sealog levels="error">
	<exceptions>
		<exception filepattern="testfile.go" minlevel="trace"/>
	</exceptions>
</sealog>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewListConstraints([]LogLevel{ErrorLvl})
		minMaxConstraint, _ := NewMinMaxConstraints(TraceLvl, CriticalLvl)
		exception, _ = NewLogLevelException("*", "testfile.go", minMaxConstraint)
		testExpected.Exceptions = []*LogLevelException{exception}
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher([]interface{}{testConsoleWriter})
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Exceptions: allowing #2"
		testConfig =
			`
<sealog levels="off">
	<exceptions>
		<exception filepattern="testfile.go" minlevel="warn"/>
	</exceptions>
</sealog>
`
		testExpected = new(LogConfig)
		testExpected.Constraints, _ = NewOffConstraints()
		minMaxConstraint, _ = NewMinMaxConstraints(WarnLvl, CriticalLvl)
		exception, _ = NewLogLevelException("*", "testfile.go", minMaxConstraint)
		testExpected.Exceptions = []*LogLevelException{exception}
		testConsoleWriter, _ = writers.NewConsoleWriter()
		testHeadSplitter, _ = dispatchers.NewSplitDispatcher([]interface{}{testConsoleWriter})
		testExpected.RootDispatcher = testHeadSplitter
		parserTests = append(parserTests, parserTest{testName, testConfig, testExpected, false})

		testName = "Errors #11"
		testConfig = `
<sealog><exceptions>
		<exception filepattern="testfile.go" minlevel="trace"/>
		<exception filepattern="testfile.go" minlevel="warn"/>
</exceptions></sealog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #12"
		testConfig = `
<sealog><exceptions>
		<exception filepattern="!@+$)!!%&@(^$" minlevel="trace"/>
</exceptions></sealog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #13"
		testConfig = `
<sealog><exceptions>
		<exception filepattern="*" minlevel="unknown"/>
</exceptions></sealog>`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #14"
		testConfig = `
<sealog levels=”off”>
	<exceptions>
		<exception filepattern="testfile.go" minlevel="off"/>
	</exceptions>
</sealog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #15"
		testConfig = `
<sealog levels=”trace”>
	<exceptions>
		<exception filepattern="testfile.go" levels="trace"/>
	</exceptions>
</sealog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #16"
		testConfig = `
<sealog minlevel=”trace”>
	<exceptions>
		<exception filepattern="testfile.go" minlevel="trace"/>
	</exceptions>
</sealog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #17"
		testConfig = `
<sealog minlevel=”trace”>
	<exceptions>
		<exception filepattern="testfile.go" minlevel="warn"/>
	</exceptions>
	<exceptions>
		<exception filepattern="testfile.go" minlevel="warn"/>
	</exceptions>
</sealog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #18"
		testConfig = `
<sealog minlevel=”trace”>
	<exceptions>
		<exception filepattern="testfile.go"/>
	</exceptions>
</sealog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #19"
		testConfig = `
<sealog minlevel=”trace”>
	<exceptions>
		<exception minlevel="warn"/>
	</exceptions>
</sealog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})

		testName = "Errors #20"
		testConfig = `
<sealog minlevel=”trace”>
	<exceptions>
		<exception/>
	</exceptions>
</sealog>
`
		parserTests = append(parserTests, parserTest{testName, testConfig, nil, true})
	}

	return parserTests
}

func TestParser(t *testing.T) {

	testFSWrapper, err := test.NewEmptyFSTestWrapper()

	if err != nil {
		t.Fatalf("Fatal error in test fs initialization: %s", err.String())
	}

	writers.SetTestMode(testFSWrapper)

	for _, test := range getParserTests() {

		conf, err := ConfigFromReader(strings.NewReader(test.config))

		if (err != nil) != test.errorExpected {
			t.Errorf("\n----ERROR in %s:\nConfig: %s\n* Expected error:%t. Got error: %t\n", test.testName,
				test.config, test.errorExpected, (err != nil))
			if err != nil {
				t.Logf("%s\n", err.String())
			}
			continue
		}

		if err == nil && !reflect.DeepEqual(conf, test.expected) {
			t.Errorf("\n----ERROR in %s:\nConfig: %s\n* Expected: %s. \n* Got: %s\n", test.testName, test.config, test.expected, conf)
		}
	}
}
