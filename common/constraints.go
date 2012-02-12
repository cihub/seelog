// Copyright (c) 2012 - Cloud Instruments Co. Ltd.
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

package common

import (
	"errors"
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
func NewMinMaxConstraints(min LogLevel, max LogLevel) (*MinMaxConstraints, error) {
	if min > max {
		return nil, errors.New(fmt.Sprintf("Min level can't be greater than max. Got min: %d, max: %d", min, max))
	}
	if min < TraceLvl || min > CriticalLvl {
		return nil, errors.New(fmt.Sprintf("Min level can't be less than Trace or greater than Critical. Got min: %d", min))
	}
	if max < TraceLvl || max > CriticalLvl {
		return nil, errors.New(fmt.Sprintf("Max level can't be less than Trace or greater than Critical. Got max: %d", max))
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
func NewListConstraints(allowList []LogLevel) (*ListConstraints, error) {
	if allowList == nil {
		return nil, errors.New("List can't be nil")
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

func createMapFromList(allowedList []LogLevel) (map[LogLevel]bool, error) {
	allowedLevels := make(map[LogLevel]bool, 0)
	for _, level := range allowedList {
		if level < TraceLvl || level > Off {
			return nil, errors.New(fmt.Sprintf("Level can't be less than Trace or greater than Critical. Got level: %d", level))
		}
		allowedLevels[level] = true
	}
	return allowedLevels, nil
}
func validateOffLevel(allowedLevels map[LogLevel]bool) error {
	if _, ok := allowedLevels[Off]; ok && len(allowedLevels) > 1 {
		return errors.New("LogLevel Off cant be mixed with other levels")
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

func NewOffConstraints() (*OffConstraints, error) {
	return &OffConstraints{}, nil
}

func (offConstr *OffConstraints) IsAllowed(level LogLevel) bool {
	return false
}

func (offConstr *OffConstraints) String() string {
	return "Off constraint"
}
