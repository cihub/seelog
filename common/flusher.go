// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

// FlusherInterface represents all objects that have to do cleanup
// at certain moments of time (e.g. before app shutdown to avoid data loss)
type FlusherInterface interface {
	Flush()
}