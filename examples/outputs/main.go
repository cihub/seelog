// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	log "github.com/cihub/sealog"
	"time"
)

func main() {
	defer log.Flush()
	runExample(consoleWriter)
	runExample(fileWriter)
	runExample(rollingFileWriter)
	runExample(rollingFileWriterManyRolls)
	runExample(bufferedWriter)
	runExample(bufferedWriterWithOverflow)
	runExample(splitDispatcher)
	runExample(filterDispatcher)
}

func runExample(exampleFunc func()) {
	exampleFunc()
	fmt.Println()
}

func consoleWriter() {
	fmt.Println("Console writer")
	
	testConfig := `
<sealog>
	<outputs>
		<console />
	</outputs>
</sealog>
`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	doLog()
}

func fileWriter() {
	fmt.Println("File writer")
	
	testConfig := `
<sealog>
	<outputs>
		<file path="log.log"/>
	</outputs>
</sealog>
`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	doLog()
}

func rollingFileWriter() {
	fmt.Println("Rolling file writer")
	
	testConfig := `
<sealog>
	<outputs>
		<rollingfile type="size" filename="roll.log" maxsize="100" maxrolls="5" />
	</outputs>
</sealog>
`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	doLog()
}

func rollingFileWriterManyRolls() {
	fmt.Println("Rolling file writer. Many rolls")
	
	testConfig := `
<sealog>
	<outputs>
		<rollingfile type="size" filename="manyrolls.log" maxsize="100" maxrolls="4" />
	</outputs>
</sealog>
`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	doLogBig()
}

func bufferedWriter() {
	fmt.Println("Buffered file writer")
	
	testConfig := `
<sealog>
	<outputs>
		<buffered size="10000" flushperiod="1000">
			<file path="bufFile.log"/>
		</buffered>
	</outputs>
</sealog>
`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	for i := 0; i < 3; i++ {
		doLog()	
		time.Sleep(1e9)
	}
	
	time.Sleep(2e9)
}

func bufferedWriterWithOverflow() {
	fmt.Println("Buffered file writer with overflow")
	
	testConfig := `
<sealog>
	<outputs>
		<buffered size="20">
			<file path="bufOverflow.log"/>
		</buffered>
	</outputs>
</sealog>
`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)

	for i := 0; i < 3; i++ {
		doLog()	
		time.Sleep(1e9)
	}
	
	time.Sleep(1e9)
}


func splitDispatcher() {
	fmt.Println("Split dispatcher")
	
	testConfig := `
<sealog>
	<outputs>
		<file path="split.log"/>
		<console />
	</outputs>
</sealog>
`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	doLog()	
}

func filterDispatcher() {
	fmt.Println("Filter dispatcher")
	
	testConfig := `
<sealog>
	<outputs>
		<filter levels="trace">
			<file path="filter.log"/>
		</filter>
		<console />
	</outputs>
</sealog>
`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	
	for i:=0; i < 5; i++ {
		log.Trace("This message on console and in file")
		log.Debug("This message only on console")
	}
}

func doLog() {
	for i:=0; i < 5; i++ {
		log.Trace("%d", i)
	}
}

func doLogBig() {
	for i:=0; i < 50; i++ {
		log.Trace("%d", i)
	}
}