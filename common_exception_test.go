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
	"testing"
)

type exceptionTestCase struct {
	funcPattern string
	filePattern string
	funcName    string
	fileName    string
	match       bool
}

var exceptionTestCases = []exceptionTestCase{
	exceptionTestCase{"*", "*", "func", "file", true},
	exceptionTestCase{"func*", "*", "func", "file", true},
	exceptionTestCase{"*func", "*", "func", "file", true},
	exceptionTestCase{"*func", "*", "1func", "file", true},
	exceptionTestCase{"func*", "*", "func1", "file", true},
	exceptionTestCase{"fu*nc", "*", "func", "file", true},
	exceptionTestCase{"fu*nc", "*", "fu1nc", "file", true},
	exceptionTestCase{"fu*nc", "*", "func1nc", "file", true},
	exceptionTestCase{"*fu*nc*", "*", "somefuntonc", "file", true},
	exceptionTestCase{"fu*nc", "*", "f1nc", "file", false},
	exceptionTestCase{"func*", "*", "fun", "file", false},
	exceptionTestCase{"fu*nc", "*", "func1n", "file", false},
	exceptionTestCase{"**f**u**n**c**", "*", "func1n", "file", true},
}

func TestMatchingCorrectness(t *testing.T) {
	constraints, err := newListConstraints([]LogLevel{TraceLvl})
	if err != nil {
		t.Error(err)
		return
	}

	for _, testCase := range exceptionTestCases {
		rule, ruleError := newLogLevelException(testCase.funcPattern, testCase.filePattern, constraints)
		if ruleError != nil {
			t.Fatalf("Unexpected error on rule creation: [ %v, %v ]. %v",
				testCase.funcPattern, testCase.filePattern, ruleError)
		}

		match := rule.match(testCase.funcName, testCase.fileName)
		if match != testCase.match {
			t.Errorf("Incorrect matching for [ %v, %v ] [ %v, %v ] Expected: %t. Got: %t",
				testCase.funcPattern, testCase.filePattern, testCase.funcName, testCase.fileName, testCase.match, match)
		}
	}
}

func TestAsterisksReducing(t *testing.T) {
	constraints, err := newListConstraints([]LogLevel{TraceLvl})
	if err != nil {
		t.Error(err)
		return
	}

	rule, err := newLogLevelException("***func**", "fi*****le", constraints)
	if err != nil {
		t.Error(err)
		return
	}
	expectFunc := "*func*"
	if rule.FuncPattern() != expectFunc {
		t.Errorf("Asterisks must be reduced. Expect:%v, Got:%v", expectFunc, rule.FuncPattern())
	}

	expectFile := "fi*le"
	if rule.FilePattern() != expectFile {
		t.Errorf("Asterisks must be reduced. Expect:%v, Got:%v", expectFile, rule.FilePattern())
	}
}
