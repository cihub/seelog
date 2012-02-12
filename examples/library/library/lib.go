// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testlibrary

import (
)

// calculateF is a meaningless example which just imitates some 
// heavy calculation operation and performs logging inside
func CalculateF(x, y int) int {
	logger.Info("Calculating F(%d, %d)", x, y)
	
	for i := 0; i < 10; i++ {
		logger.Trace("F calc iteration %d", i)
	}
	
	result := x + y
	
	logger.Debug("F = %d", result)
	return result
}