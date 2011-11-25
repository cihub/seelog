package common

import (
	"testing"
)

func TestInvalidMinMaxConstraints(t *testing.T) {
	constr, err := NewMinMaxConstraints(CriticalLvl, WarnLvl)

	if err == nil || constr != nil {
		t.Errorf("Expected an error and a nil value for minmax constraints: min = %d, max = %d. Got: %v, %v",
			CriticalLvl, WarnLvl, err, constr)
		return
	}
}

func TestInvalidLogLevels(t *testing.T) {
	var invalidMin uint8 = 123
	var invalidMax uint8 = 124
	minMaxConstr, errMinMax := NewMinMaxConstraints(LogLevel(invalidMin), LogLevel(invalidMax))

	if errMinMax == nil || minMaxConstr != nil {
		t.Errorf("Expected an error and a nil value for minmax constraints: min = %d, max = %d. Got: %v, %v",
			invalidMin, invalidMax, errMinMax, minMaxConstr)
		return
	}

	invalidList := []LogLevel{145}

	listConstr, errList := NewListConstraints(invalidList)

	if errList == nil || listConstr != nil {
		t.Errorf("Expected an error and a nil value for constraints list: %v. Got: %v, %v",
			invalidList, errList, listConstr)
		return
	}
}

func TestListConstraintsWithDuplicates(t *testing.T) {
	duplicateList := []LogLevel{TraceLvl, DebugLvl, InfoLvl,
		WarnLvl, ErrorLvl, CriticalLvl, CriticalLvl, CriticalLvl}

	listConstr, errList := NewListConstraints(duplicateList)

	if errList != nil || listConstr == nil {
		t.Errorf("Expected a valid constraints list struct for: %v, got error: %v, value: %v",
			duplicateList, errList, listConstr)
		return
	}

	listLevels := listConstr.AllowedLevels()

	if listLevels == nil {
		t.Fatalf("listConstr.AllowedLevels() == nil")
		return
	}

	if len(listLevels) != 6 {
		t.Errorf("Expected: listConstr.AllowedLevels() length == 6. Got: %d", len(listLevels))
		return
	}
}

func TestListConstraintsWithOffInList(t *testing.T) {
	offList := []LogLevel{TraceLvl, DebugLvl, Off}

	listConstr, errList := NewListConstraints(offList)

	if errList == nil || listConstr != nil {
		t.Errorf("Expected an error and a nil value for constraints list with 'Off':  %v. Got: %v, %v",
			offList, errList, listConstr)
		return
	}
}

type logLevelTestCase struct {
	level   LogLevel
	allowed bool
}

var minMaxTests = []logLevelTestCase{
	logLevelTestCase{TraceLvl, false},
	logLevelTestCase{DebugLvl, false},
	logLevelTestCase{InfoLvl, true},
	logLevelTestCase{WarnLvl, true},
	logLevelTestCase{ErrorLvl, false},
	logLevelTestCase{CriticalLvl, false},
	logLevelTestCase{123, false},
	logLevelTestCase{6, false},
}

func TestValidMinMaxConstraints(t *testing.T) {

	constr, err := NewMinMaxConstraints(InfoLvl, WarnLvl)

	if err != nil || constr == nil {
		t.Errorf("Expected a valid constraints struct for minmax constraints: min = %d, max = %d. Got: %v, %v",
			InfoLvl, WarnLvl, err, constr)
		return
	}

	for _, minMaxTest := range minMaxTests {
		allowed := constr.IsAllowed(minMaxTest.level)
		if allowed != minMaxTest.allowed {
			t.Errorf("Expected IsAllowed() = %t for level = %d. Got: %t",
				minMaxTest.allowed, minMaxTest.level, allowed)
			return
		}
	}
}

var listTests = []logLevelTestCase{
	logLevelTestCase{TraceLvl, true},
	logLevelTestCase{DebugLvl, false},
	logLevelTestCase{InfoLvl, true},
	logLevelTestCase{WarnLvl, true},
	logLevelTestCase{ErrorLvl, false},
	logLevelTestCase{CriticalLvl, true},
	logLevelTestCase{123, false},
	logLevelTestCase{6, false},
}

func TestValidListConstraints(t *testing.T) {
	validList := []LogLevel{TraceLvl, InfoLvl, WarnLvl, CriticalLvl}
	constr, err := NewListConstraints(validList)

	if err != nil || constr == nil {
		t.Errorf("Expected a valid constraints list struct for: %v. Got error: %v, value: %v",
			validList, err, constr)
		return
	}

	for _, minMaxTest := range listTests {
		allowed := constr.IsAllowed(minMaxTest.level)
		if allowed != minMaxTest.allowed {
			t.Errorf("Expected IsAllowed() = %t for level = %d. Got: %t",
				minMaxTest.allowed, minMaxTest.level, allowed)
			return
		}
	}
}

var offTests = []logLevelTestCase{
	logLevelTestCase{TraceLvl, false},
	logLevelTestCase{DebugLvl, false},
	logLevelTestCase{InfoLvl, false},
	logLevelTestCase{WarnLvl, false},
	logLevelTestCase{ErrorLvl, false},
	logLevelTestCase{CriticalLvl, false},
	logLevelTestCase{123, false},
	logLevelTestCase{6, false},
}

func TestValidListOffConstraints(t *testing.T) {
	validList := []LogLevel{Off}
	constr, err := NewListConstraints(validList)

	if err != nil || constr == nil {
		t.Errorf("Expected a valid constraints list struct for: %v. Got error: %v, value: %v",
			validList, err, constr)
		return
	}

	for _, minMaxTest := range offTests {
		allowed := constr.IsAllowed(minMaxTest.level)
		if allowed != minMaxTest.allowed {
			t.Errorf("Expected IsAllowed() = %t for level = %d. Got: %t",
				minMaxTest.allowed, minMaxTest.level, allowed)
			return
		}
	}
}
