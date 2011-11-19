package sealog

import (
	"testing"
	"sealog/common"
)

func TestInvalidMinMaxConstraints(t *testing.T) {
	constr, err := NewMinMaxConstraints(common.CriticalLvl, common.WarnLvl)

	if err == nil || constr != nil {
		t.Errorf("Expected an error and a nil value for minmax constraints: min = %d, max = %d. Got: %v, %v",
			common.CriticalLvl, common.WarnLvl, err, constr)
		return
	}
}

func TestInvalidLogLevels(t *testing.T) {
	var invalidMin uint8 = 123
	var invalidMax uint8 = 124
	minMaxConstr, errMinMax := NewMinMaxConstraints(common.LogLevel(invalidMin), common.LogLevel(invalidMax))

	if errMinMax == nil || minMaxConstr != nil {
		t.Errorf("Expected an error and a nil value for minmax constraints: min = %d, max = %d. Got: %v, %v",
			invalidMin, invalidMax, errMinMax, minMaxConstr)
		return
	}

	invalidList := []common.LogLevel{145}

	listConstr, errList := NewListConstraints(invalidList)

	if errList == nil || listConstr != nil {
		t.Errorf("Expected an error and a nil value for constraints list: %v. Got: %v, %v",
			invalidList, errList, listConstr)
		return
	}
}

func TestListConstraintsWithDuplicates(t *testing.T) {
	duplicateList := []common.LogLevel{common.TraceLvl, common.DebugLvl, common.InfoLvl,
		common.WarnLvl, common.ErrorLvl, common.CriticalLvl, common.CriticalLvl, common.CriticalLvl}

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
	offList := []common.LogLevel{common.TraceLvl, common.DebugLvl, common.Off}

	listConstr, errList := NewListConstraints(offList)

	if errList == nil || listConstr != nil {
		t.Errorf("Expected an error and a nil value for constraints list with 'Off':  %v. Got: %v, %v",
			offList, errList, listConstr)
		return
	}
}


type logLevelTestCase struct {
	level   common.LogLevel
	allowed bool
}

var minMaxTests = []logLevelTestCase {
	logLevelTestCase{common.TraceLvl, false},
	logLevelTestCase{common.DebugLvl, false},
	logLevelTestCase{common.InfoLvl, true},
	logLevelTestCase{common.WarnLvl, true},
	logLevelTestCase{common.ErrorLvl, false},
	logLevelTestCase{common.CriticalLvl, false},
	logLevelTestCase{123, false},
	logLevelTestCase{6, false},
}

func TestValidMinMaxConstraints(t *testing.T) {

	constr, err := NewMinMaxConstraints(common.InfoLvl, common.WarnLvl)

	if err != nil || constr == nil {
		t.Errorf("Expected a valid constraints struct for minmax constraints: min = %d, max = %d. Got: %v, %v",
			common.InfoLvl, common.WarnLvl, err, constr)
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

var listTests = []logLevelTestCase {
	logLevelTestCase{common.TraceLvl, true},
	logLevelTestCase{common.DebugLvl, false},
	logLevelTestCase{common.InfoLvl, true},
	logLevelTestCase{common.WarnLvl, true},
	logLevelTestCase{common.ErrorLvl, false},
	logLevelTestCase{common.CriticalLvl, true},
	logLevelTestCase{123, false},
	logLevelTestCase{6, false},
}

func TestValidListConstraints(t *testing.T) {
	validList := []common.LogLevel{common.TraceLvl, common.InfoLvl, common.WarnLvl, common.CriticalLvl}
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

var offTests = []logLevelTestCase {
	logLevelTestCase{common.TraceLvl, false},
	logLevelTestCase{common.DebugLvl, false},
	logLevelTestCase{common.InfoLvl, false},
	logLevelTestCase{common.WarnLvl, false},
	logLevelTestCase{common.ErrorLvl, false},
	logLevelTestCase{common.CriticalLvl, false},
	logLevelTestCase{123, false},
	logLevelTestCase{6, false},
}

func TestValidListOffConstraints(t *testing.T) {
	validList := []common.LogLevel{common.Off}
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
