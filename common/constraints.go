package common

import (
	"os"
	"fmt"
	"strings"
)

// Represents constraints which form a general rule for log levels selection
type LogLevelConstraints interface {
	IsAllowed(level LogLevel) bool
}

// A MinMaxConstraints represents constraints which use minimal and maximal allowed log levels.
type MinMaxConstraints struct {
	min LogLevel
	max LogLevel
}

// NewMinMaxConstraints creates a new MinMaxConstraints struct with the specified min and max levels.
func NewMinMaxConstraints(min LogLevel, max LogLevel) (*MinMaxConstraints, os.Error) {
	if min > max {
		return nil, os.NewError(fmt.Sprintf("Min level can't be greater than max. Got min: %d, max: %d", min, max))
	}
	if min < TraceLvl || min > CriticalLvl {
		return nil, os.NewError(fmt.Sprintf("Min level can't be less than Trace or greater than Critical. Got min: %d", min))
	}
	if max < TraceLvl || max > CriticalLvl {
		return nil, os.NewError(fmt.Sprintf("Max level can't be less than Trace or greater than Critical. Got max: %d", max))
	}

	return &MinMaxConstraints{min, max}, nil
}

// IsAllowed returns true, if log level is in [min, max] range (inclusive).
func (minMaxConstr *MinMaxConstraints) IsAllowed(level LogLevel) bool {
	return level >= minMaxConstr.min && level <= minMaxConstr.max
}

func (minMaxConstr *MinMaxConstraints) String() string {
	return fmt.Sprintf("Min: %s. Max: %s", minMaxConstr.min, minMaxConstr.max)
}

//=======================================================

// A ListConstraints represents constraints which use allowed log levels list.
type ListConstraints struct {
	allowedLevels map[LogLevel]bool
}

// NewListConstraints creates a new ListConstraints struct with the specified allowed levels.
func NewListConstraints(allowList []LogLevel) (*ListConstraints, os.Error) {
	if allowList == nil {
		return nil, os.NewError("List can't be nil")
	}

	allowLevels, err := createMapFromList(allowList)
	if err != nil {
		return nil, err
	}
	err = validateOffLevel(allowLevels)
	if err != nil {
		return nil, err
	}

	return &ListConstraints{allowLevels}, nil
}

func (listConstr *ListConstraints) String() string {
	allowedList := "List: "

	listLevel := make([]string, len(listConstr.allowedLevels))

	i := 0
	for level, _ := range listConstr.allowedLevels {
		listLevel[i] = level.String()
		i++
	}

	allowedList += strings.Join(listLevel, ",")

	return allowedList
}

func createMapFromList(allowedList []LogLevel) (map[LogLevel]bool, os.Error) {
	allowedLevels := make(map[LogLevel]bool, 0)
	for _, level := range allowedList {
		if level < TraceLvl || level > Off {
			return nil, os.NewError(fmt.Sprintf("Level can't be less than Trace or greater than Critical. Got level: %d", level))
		}
		allowedLevels[level] = true
	}
	return allowedLevels, nil
}
func validateOffLevel(allowedLevels map[LogLevel]bool) os.Error {
	if _, ok := allowedLevels[Off]; ok && len(allowedLevels) > 1 {
		return os.NewError("LogLevel Off cant be mixed with other levels")
	}

	return nil
}

// IsAllowed returns true, if log level is in allowed log levels list.
// If the list contains the only item 'common.Off' then IsAllowed will always return false for any input values.
func (listConstr *ListConstraints) IsAllowed(level LogLevel) bool {
	for l, _ := range listConstr.allowedLevels {
		if l == level && level != Off {
			return true
		}
	}

	return false
}

// AllowedLevels returns allowed levels configuration as a map.
func (listConstr *ListConstraints) AllowedLevels() map[LogLevel]bool {
	return listConstr.allowedLevels
}

//=======================================================

type OffConstraints struct {

}

func NewOffConstraints() (*OffConstraints, os.Error) {
	return &OffConstraints{}, nil
}

func (offConstr *OffConstraints) IsAllowed(level LogLevel) bool {
	return false
}

func (offConstr *OffConstraints) String() string {
	return "Off constraint"
}
