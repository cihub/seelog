// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	. "github.com/cihub/sealog/common"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	testConfig :=
		`
<sealog levels="trace, debug">
	<exceptions>
		<exception funcpattern="*getFirst*" filepattern="*" minlevel="off" />
		<exception funcpattern="*getSecond*" filepattern="*" levels="info, error" />
	</exceptions>
</sealog>
`

	conf, err := ConfigFromReader(strings.NewReader(testConfig))
	if err != nil {
		t.Errorf("Parse error: %s\n", err.Error())
		return
	}

	context, err := CurrentContext()
	if err != nil {
		t.Errorf("Cannot get current context:" + err.Error())
		return
	}
	firstContext, err := getFirstContext()
	if err != nil {
		t.Errorf("Cannot get current context:" + err.Error())
		return
	}
	secondContext, err := getSecondContext()
	if err != nil {
		t.Errorf("Cannot get current context:" + err.Error())
		return
	}

	if !conf.IsAllowed(TraceLvl, context) {
		t.Errorf("Error: deny trace in current context")
	}
	if conf.IsAllowed(TraceLvl, firstContext) {
		t.Errorf("Error: allow trace in first context")
	}
	if conf.IsAllowed(ErrorLvl, context) {
		t.Errorf("Error: allow error in current context")
	}
	if !conf.IsAllowed(ErrorLvl, secondContext) {
		t.Errorf("Error: deny error in second context")
	}

	// cache test
	if !conf.IsAllowed(TraceLvl, context) {
		t.Errorf("Error: deny trace in current context")
	}
	if conf.IsAllowed(TraceLvl, firstContext) {
		t.Errorf("Error: allow trace in first context")
	}
	if conf.IsAllowed(ErrorLvl, context) {
		t.Errorf("Error: allow error in current context")
	}
	if !conf.IsAllowed(ErrorLvl, secondContext) {
		t.Errorf("Error: deny error in second context")
	}
}

func getFirstContext() (*LogContext, error) {
	return CurrentContext()
}

func getSecondContext() (*LogContext, error) {
	return CurrentContext()
}
