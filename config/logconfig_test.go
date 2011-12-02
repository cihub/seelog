// Package config contains configuration functionality of sealog
package config

import (
	"testing"
	. "sealog/common"
	"strings"
	"os"
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
		t.Errorf("Parse error: %s\n", err.String())
		return
	}
	
	context, err := CurrentContext()
	if err != nil {
		t.Errorf("Cannot get current context:" + err.String())
		return
	}
	firstContext, err := getFirstContext()
	if err != nil {
		t.Errorf("Cannot get current context:" + err.String())
		return
	}
	secondContext, err := getSecondContext()
	if err != nil {
		t.Errorf("Cannot get current context:" + err.String())
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

func getFirstContext() (*LogContext, os.Error) {
	return CurrentContext()
}

func getSecondContext() (*LogContext, os.Error) {
	return CurrentContext()
}