package sealog

import (
	"testing"
	"sealog/common"
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
	constraints, err := NewListConstraints([]common.LogLevel{common.TraceLvl})
	if err != nil {
		t.Error(err)
		return
	}

	for _, testCase := range exceptionTestCases {
		rule, ruleError := NewLogLevelException(testCase.funcPattern, testCase.filePattern, constraints)
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
