package sealog

import (
	"os"
	"sealog/common"
	"fmt"
)

// Represents constraints which form a general rule for log levels selection
type LogLevelConstraints interface {
	IsAllowed(level common.LogLevel) bool
}

// A MinMaxConstraints represents constraints which use minimal and maximal allowed log levels.
type MinMaxConstraints struct {
	min common.LogLevel
	max common.LogLevel
}

// NewMinMaxConstraints creates a new MinMaxConstraints struct with the specified min and max levels.
func NewMinMaxConstraints(min common.LogLevel, max common.LogLevel) (*MinMaxConstraints, os.Error) {
	if min > max {
		return nil, os.NewError(fmt.Sprintf("Min level can't be greater than max. Got min: %d, max: %d", min, max))
	}
	if min < common.TraceLvl || min > common.CriticalLvl {
		return nil, os.NewError(fmt.Sprintf("Min level can't be less than Trace or greater than Critical. Got min: %d", min))
	}
	if max < common.TraceLvl || max > common.CriticalLvl {
		return nil, os.NewError(fmt.Sprintf("Max level can't be less than Trace or greater than Critical. Got max: %d", max))
	}

	return &MinMaxConstraints{min, max}, nil
}

// IsAllowed returns true, if log level is in [min, max] range (inclusive).
func (this *MinMaxConstraints) IsAllowed(level common.LogLevel) bool {
	return level >= this.min && level <= this.max
}


// A ListConstraints represents constraints which use allowed log levels list.
type ListConstraints struct {
	allowLevels map[common.LogLevel]bool
}

// NewListConstraints creates a new ListConstraints struct with the specified allowed levels.
func NewListConstraints(allowList []common.LogLevel) (*ListConstraints, os.Error) {
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

func createMapFromList(allowList []common.LogLevel) (map[common.LogLevel]bool, os.Error) {
	allowLevels := make(map[common.LogLevel]bool, 0)
	for _, level := range allowList {
		if level < common.TraceLvl || level > common.Off {
			return nil, os.NewError(fmt.Sprintf("Level can't be less than Trace or greater than Critical. Got level: %d", level))
		}
		allowLevels[level] = true
	}
	return allowLevels, nil
}
func validateOffLevel(allowLevels map[common.LogLevel]bool) os.Error {
	if _, ok := allowLevels[common.Off]; ok && len(allowLevels) > 1 {
		return os.NewError("LogLevel Off cant be mixed with other levels")
	}

	return nil
}

// IsAllowed returns true, if log level is in allowed log levels list.
// If the list contains the only item 'common.Off' then IsAllowed will always return false for any input values.
func (this *ListConstraints) IsAllowed(level common.LogLevel) bool {
	for l, _ := range this.allowLevels {
		if l == level && level != common.Off {
			return true
		}
	}

	return false
}

// AllowedLevels returns allowed levels configuration as a map.
func (this *ListConstraints) AllowedLevels() map[common.LogLevel]bool {
	return this.allowLevels
}
