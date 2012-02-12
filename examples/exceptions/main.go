// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	log "github.com/cihub/seelog"
)

func main() {
	defer log.Flush()
	testMinMax()
	testMin()
	testMax()
	testList()
	testFuncException()
	testFileException()
}


func testMinMax() {
	fmt.Println("testMinMax")
	testConfig := `
<seelog type="sync" minlevel="info" maxlevel="error">
	<outputs><console/></outputs>
</seelog>`

	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	log.Trace("NOT Printed")
	log.Debug("NOT Printed")
	log.Info("Printed")
	log.Warn("Printed")
	log.Error("Printed")
	log.Critical("NOT Printed")
}

func testMin() {
	fmt.Println("testMin")
	testConfig := `
<seelog type="sync" minlevel="info">
	<outputs><console/></outputs>
</seelog>`

	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	log.Trace("NOT Printed")
	log.Debug("NOT Printed")
	log.Info("Printed")
	log.Warn("Printed")
	log.Error("Printed")
	log.Critical("Printed")
}

func testMax() {
	fmt.Println("testMax")
	testConfig := `
<seelog type="sync" maxlevel="error">
	<outputs><console/></outputs>
</seelog>`

	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	log.Trace("Printed")
	log.Debug("Printed")
	log.Info("Printed")
	log.Warn("Printed")
	log.Error("Printed")
	log.Critical("NOT Printed")
}

func testList() {
	fmt.Println("testList")
	testConfig := `
<seelog type="sync" levels="info, trace, critical">
	<outputs><console/></outputs>
</seelog>`

	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	log.Trace("Printed")
	log.Debug("NOT Printed")
	log.Info("Printed")
	log.Warn("NOT Printed")
	log.Error("NOT Printed")
	log.Critical("Printed")
}

func testFuncException() {
	fmt.Println("testFuncException")
	testConfig := `
<seelog type="sync" minlevel="info">
	<exceptions>
		<exception funcpattern="*main.test*Except*" minlevel="error"/>
	</exceptions>
	<outputs>
		<console/>
	</outputs>
</seelog>`

	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	log.Trace("NOT Printed")
	log.Debug("NOT Printed")
	log.Info("NOT Printed")
	log.Warn("NOT Printed")
	log.Error("Printed")
	log.Critical("Printed")
}

func testFileException() {
	fmt.Println("testFileException")
	testConfig := `
<seelog type="sync" minlevel="info">
	<exceptions>
		<exception filepattern="*main.go" minlevel="error"/>
	</exceptions>
	<outputs>
		<console/>
	</outputs>
</seelog>`

	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	log.Trace("NOT Printed")
	log.Debug("NOT Printed")
	log.Info("NOT Printed")
	log.Warn("NOT Printed")
	log.Error("Printed")
	log.Critical("Printed")
}