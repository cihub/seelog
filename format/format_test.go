// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package format

import (
	. "github.com/cihub/seelog/common"
	"strings"
	"testing"
	"time"
)

const (
	TestFuncName = "TestFormats"
)

type formatTest struct {
	formatString   string
	input          string
	inputLogLevel  LogLevel
	expectedOutput string
	errorExpected  bool
}

var formatTests = []formatTest{
	{"test", "abcdef", TraceLvl, "test", false},
	{"", "abcdef", TraceLvl, "", false},
	{"%Level", "", TraceLvl, "Trace", false},
	{"%Level", "", DebugLvl, "Debug", false},
	{"%Level", "", InfoLvl, "Info", false},
	{"%Level", "", WarnLvl, "Warn", false},
	{"%Level", "", ErrorLvl, "Error", false},
	{"%Level", "", CriticalLvl, "Critical", false},
	{"[%Level]", "", TraceLvl, "[Trace]", false},
	{"[%Level]", "abc", DebugLvl, "[Debug]", false},
	{"%LevelLevel", "", InfoLvl, "InfoLevel", false},
	{"[%Level][%Level]", "", WarnLvl, "[Warn][Warn]", false},
	{"[%Level]X[%Level]", "", ErrorLvl, "[Error]X[Error]", false},
	{"%Levelll", "", CriticalLvl, "Criticalll", false},
	{"%Lvl", "", TraceLvl, "", true},
	{"%%Level", "", DebugLvl, "%Level", false},
	{"%Level%", "", InfoLvl, "", true},
	{"%sevel", "", WarnLvl, "", true},
	{"Level", "", ErrorLvl, "Level", false},
	{"%LevelLevel", "", CriticalLvl, "CriticalLevel", false},
	{"%Lev", "", TraceLvl, "Trc", false},
	{"%Lev", "", DebugLvl, "Dbg", false},
	{"%Lev", "", InfoLvl, "Inf", false},
	{"%Lev", "", WarnLvl, "Wrn", false},
	{"%Lev", "", ErrorLvl, "Err", false},
	{"%Lev", "", CriticalLvl, "Crt", false},
	{"[%Lev]", "", TraceLvl, "[Trc]", false},
	{"[%Lev]", "abc", DebugLvl, "[Dbg]", false},
	{"%LevLevel", "", InfoLvl, "InfLevel", false},
	{"[%Level][%Lev]", "", WarnLvl, "[Warn][Wrn]", false},
	{"[%Lev]X[%Lev]", "", ErrorLvl, "[Err]X[Err]", false},
	{"%Levll", "", CriticalLvl, "Crtll", false},
	{"%LEVEL", "", TraceLvl, "TRACE", false},
	{"%LEVEL", "", DebugLvl, "DEBUG", false},
	{"%LEVEL", "", InfoLvl, "INFO", false},
	{"%LEVEL", "", WarnLvl, "WARN", false},
	{"%LEVEL", "", ErrorLvl, "ERROR", false},
	{"%LEVEL", "", CriticalLvl, "CRITICAL", false},
	{"[%LEVEL]", "", TraceLvl, "[TRACE]", false},
	{"[%LEVEL]", "abc", DebugLvl, "[DEBUG]", false},
	{"%LEVELLEVEL", "", InfoLvl, "INFOLEVEL", false},
	{"[%LEVEL][%LEVEL]", "", WarnLvl, "[WARN][WARN]", false},
	{"[%LEVEL]X[%Level]", "", ErrorLvl, "[ERROR]X[Error]", false},
	{"%LEVELLL", "", CriticalLvl, "CRITICALLL", false},
	{"%LEV", "", TraceLvl, "TRC", false},
	{"%LEV", "", DebugLvl, "DBG", false},
	{"%LEV", "", InfoLvl, "INF", false},
	{"%LEV", "", WarnLvl, "WRN", false},
	{"%LEV", "", ErrorLvl, "ERR", false},
	{"%LEV", "", CriticalLvl, "CRT", false},
	{"[%LEV]", "", TraceLvl, "[TRC]", false},
	{"[%LEV]", "abc", DebugLvl, "[DBG]", false},
	{"%LEVLEVEL", "", InfoLvl, "INFLEVEL", false},
	{"[%LEVEL][%LEV]", "", WarnLvl, "[WARN][WRN]", false},
	{"[%LEV]X[%LEV]", "", ErrorLvl, "[ERR]X[ERR]", false},
	{"%LEVLL", "", CriticalLvl, "CRTLL", false},
	{"%l", "", TraceLvl, "t", false},
	{"%l", "", DebugLvl, "d", false},
	{"%l", "", InfoLvl, "i", false},
	{"%l", "", WarnLvl, "w", false},
	{"%l", "", ErrorLvl, "e", false},
	{"%l", "", CriticalLvl, "c", false},
	{"[%l]", "", TraceLvl, "[t]", false},
	{"[%l]", "abc", DebugLvl, "[d]", false},
	{"%Level%Msg", "", TraceLvl, "Trace", false},
	{"%Level%Msg", "A", DebugLvl, "DebugA", false},
	{"%Level%Msg", "", InfoLvl, "Info", false},
	{"%Level%Msg", "test", WarnLvl, "Warntest", false},
	{"%Level%Msg", " ", ErrorLvl, "Error ", false},
	{"%Level%Msg", "", CriticalLvl, "Critical", false},
	{"[%Level]", "", TraceLvl, "[Trace]", false},
	{"[%Level]", "abc", DebugLvl, "[Debug]", false},
	{"%Level%MsgLevel", "A", InfoLvl, "InfoALevel", false},
	{"[%Level]%Msg[%Level]", "test", WarnLvl, "[Warn]test[Warn]", false},
	{"[%Level]%MsgX[%Level]", "test", ErrorLvl, "[Error]testX[Error]", false},
	{"%Levell%Msgl", "Test", CriticalLvl, "CriticallTestl", false},
	{"%Lev%Msg%LEVEL%LEV%l%Msg", "Test", InfoLvl, "InfTestINFOINFiTest", false},
	{"%n", "", CriticalLvl, "\n", false},
	{"%t", "", CriticalLvl, "\t", false},
}

func TestFormats(t *testing.T) {

	context, conErr := CurrentContext()
	if conErr != nil {
		t.Fatal("Cannot get current context:" + conErr.Error())
		return
	}

	for _, test := range formatTests {

		form, err := NewFormatter(test.formatString)

		if (err != nil) != test.errorExpected {
			t.Errorf("Input: %s \nInput LL: %s\n* Expected error:%t Got error: %t\n",
				test.input, test.inputLogLevel, test.errorExpected, (err != nil))
			if err != nil {
				t.Logf("%s\n", err.Error())
			}
			continue
		} else if err != nil {
			continue
		}

		msg := form.Format(test.input, test.inputLogLevel, context)

		if err == nil && msg != test.expectedOutput {
			t.Errorf("Input: %s \nInput LL: %s\n* Expected: %s \n* Got: %s\n", test.input,
				test.inputLogLevel, test.expectedOutput, msg)
		}
	}
}

func TestDateFormat(t *testing.T) {
	_, err := NewFormatter("%Date")
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
	}
}

func TestDateParametrizedFormat(t *testing.T) {
	testFormat := "Mon Jan 02 2006 15:04:05"
	preciseForamt := "Mon Jan 02 2006 15:04:05.000"

	context, conErr := CurrentContext()
	if conErr != nil {
		t.Fatal("Cannot get current context:" + conErr.Error())
		return
	}

	form, err := NewFormatter("%Date(" + preciseForamt + ")")
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
	}

	dateBefore := time.Now().Format(testFormat)
	msg := form.Format("", TraceLvl, context)
	dateAfter := time.Now().Format(testFormat)

	if !strings.HasPrefix(msg, dateBefore) && !strings.HasPrefix(msg, dateAfter) {
		t.Errorf("Incorrect message: %v. Expected %v or %v", msg, dateBefore, dateAfter)
	}
}
