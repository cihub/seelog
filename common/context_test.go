package common

import (
	"testing"
	"os"
	"path/filepath"
)

const (
	shortPath    = "context_test.go"
	commonPrefix = "sealog/common."
)

var testFullPath string

func fullPath(t *testing.T) string {
	if testFullPath == "" {
		wd, err := os.Getwd()

		if err != nil {
			t.Fatalf("Cannot get working directory: %s", err.String())
		}

		testFullPath = filepath.Join(wd, shortPath)
	}

	return testFullPath
}

func TestContext(t *testing.T) {
	context, err := CurrentContext()

	nameFunc := commonPrefix + "TestContext"

	if err != nil {
		t.Fatalf("Unexpected error: %s", err.String())
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

func innerContext() (context *LogContext, err os.Error) {
	return CurrentContext()
}

func TestInnerContext(t *testing.T) {
	context, err := innerContext()

	nameFunc := commonPrefix + "innerContext"

	if err != nil {
		t.Fatalf("Unexpected error: %s", err.String())
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
