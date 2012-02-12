// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"testing"
	"os"
	"path/filepath"
)

const (
	shortPath    = "context_test.go"
	commonPrefix = "github.com/cihub/seelog/common."
)

var testFullPath string

func fullPath(t *testing.T) string {
	if testFullPath == "" {
		wd, err := os.Getwd()

		if err != nil {
			t.Fatalf("Cannot get working directory: %s", err.Error())
		}

		testFullPath = filepath.Join(wd, shortPath)
	}

	return testFullPath
}

func TestContext(t *testing.T) {
	context, err := CurrentContext()

	nameFunc := commonPrefix + "TestContext"

	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	if context == nil {
		t.Fatalf("Expected: context != nil")
	}

	if context.Func() != nameFunc {
		t.Errorf("Expected context.Func == %s ; got %s", nameFunc, context.Func())
	}

	if context.ShortPath() != shortPath {
		t.Errorf("Expected context.ShortPath == %s ; got %s", shortPath, context.ShortPath())
	}

	fp := fullPath(t)

	if context.FullPath() != fp {
		t.Errorf("Expected context.FullPath == %s ; got %s", fp, context.FullPath())
	}
}

func innerContext() (context *LogContext, err error) {
	return CurrentContext()
}

func TestInnerContext(t *testing.T) {
	context, err := innerContext()

	nameFunc := commonPrefix + "innerContext"

	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	if context == nil {
		t.Fatalf("Expected: context != nil")
	}

	if context.Func() != nameFunc {
		t.Errorf("Expected context.Func == %s ; got %s", nameFunc, context.Func())
	}

	if context.ShortPath() != shortPath {
		t.Errorf("Expected context.ShortPath == %s ; got %s", shortPath, context.ShortPath())
	}

	fp := fullPath(t)

	if context.FullPath() != fp {
		t.Errorf("Expected context.FullPath == %s ; got %s", fp, context.FullPath())
	}
}
