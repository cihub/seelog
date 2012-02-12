// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	log "github.com/cihub/sealog"
	"github.com/cihub/sealog/test"
	"crypto/rand"
	"path/filepath"
	"math/big"
	"sync"
	"fmt"
	"os"
	"time"
)

const (
	LogDir = "log"
	goroutinesCount = 1000
	logsPerGoroutineCount = 100
	LogFile = "log.log"
)

var loggerReplacements int

var counterMutex *sync.Mutex
var waitGroup *sync.WaitGroup

var counter int64


var fileConfig = `
<sealog type="asyncloop">
	<outputs>
		<file path="` + filepath.Join(LogDir, LogFile) + `" formatid="testFormat"/>
	</outputs>
	<formats>
	    <format id="testFormat" format="%Msg%n"/>
	</formats>
</sealog>`

var fileAsyncLoopConfig = `
<sealog type="asyncloop">
	<outputs>
		<file path="` + filepath.Join(LogDir, LogFile) + `" formatid="testFormat"/>
	</outputs>
	<formats>
	    <format id="testFormat" format="%Msg%n"/>
	</formats>
</sealog>`

var fileAsyncTimer100Config = `
<sealog type="sync">
	<outputs>
		<file path="` + filepath.Join(LogDir, LogFile) + `" formatid="testFormat"/>
	</outputs>
	<formats>
	    <format id="testFormat" format="%Msg%n"/>
	</formats>
</sealog>`

var fileAsyncTimer1000Config = `
<sealog type="asynctimer" asyncinterval="1000">
	<outputs>
		<file path="` + filepath.Join(LogDir, LogFile) + `" formatid="testFormat"/>
	</outputs>
	<formats>
	    <format id="testFormat" format="%Msg%n"/>
	</formats>
</sealog>`

var fileAsyncTimer10000Config = `
<sealog type="asynctimer" asyncinterval="10000">
	<outputs>
		<file path="` + filepath.Join(LogDir, LogFile) + `" formatid="testFormat"/>
	</outputs>
	<formats>
	    <format id="testFormat" format="%Msg%n"/>
	</formats>
</sealog>`



var fileBufferedConfig = `
<sealog type="sync">
	<outputs>
		<buffered size="100" formatid="testFormat">
			<file path="` + filepath.Join(LogDir, LogFile) + `"/>
		</buffered>
	</outputs>
	<formats>
	    <format id="testFormat" format="%Msg%n"/>
	</formats>
</sealog>`

var fileBufferedAsyncLoopConfig = `
<sealog type="asyncloop">
	<outputs>
		<buffered size="100" formatid="testFormat">
			<file path="` + filepath.Join(LogDir, LogFile) + `"/>
		</buffered>
	</outputs>
	<formats>
	    <format id="testFormat" format="%Msg%n"/>
	</formats>
</sealog>`

var fileBufferedAsyncTimer100Config = `
<sealog type="asynctimer" asyncinterval="100">
	<outputs>
		<buffered size="100" formatid="testFormat">
			<file path="` + filepath.Join(LogDir, LogFile) + `"/>
		</buffered>
	</outputs>
	<formats>
	    <format id="testFormat" format="%Msg%n"/>
	</formats>
</sealog>`

var fileBufferedAsyncTimer1000Config = `
<sealog type="asynctimer" asyncinterval="1000">
	<outputs>
		<buffered size="100" formatid="testFormat">
			<file path="` + filepath.Join(LogDir, LogFile) + `"/>
		</buffered>
	</outputs>
	<formats>
	    <format id="testFormat" format="%Msg%n"/>
	</formats>
</sealog>`

var fileBufferedAsyncTimer10000Config = `
<sealog type="asynctimer" asyncinterval="10000">
	<outputs>
		<buffered size="100"  formatid="testFormat">
			<file path="` + filepath.Join(LogDir, LogFile) + `"/>
		</buffered>
	</outputs>
	<formats>
	    <format id="testFormat" format="%Msg%n"/>
	</formats>
</sealog>`


var configPool = []string {
	fileConfig,
	fileAsyncLoopConfig,
	fileAsyncTimer100Config,
	fileAsyncTimer1000Config,
	fileAsyncTimer10000Config,
	fileBufferedConfig,
	fileBufferedAsyncLoopConfig,
	fileBufferedAsyncTimer100Config,
	fileBufferedAsyncTimer1000Config,
	fileBufferedAsyncTimer10000Config,
}

func switchToRandomConfigFromPool() {
	
	configIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(configPool))))
	
	if err != nil {
		panic(fmt.Sprintf("Error during random index generation: %s", err.Error()))
	}
	
	randomCfg := configPool[int(configIndex.Int64())]
	
	logger, err := log.LoggerFromConfigAsBytes([]byte(randomCfg))

	if err != nil {
		panic(fmt.Sprintf("Error during config creation: %s", err.Error()))
	}

	log.ReplaceLogger(logger)
	loggerReplacements++
}

func logRoutine(ind int) {
	for i := 0; i < logsPerGoroutineCount; i++ {
		counterMutex.Lock()
		log.Debug("%d", counter)
		//fmt.Printf("log #%v from #%v\n", i, ind)
		counter++
		switchToRandomConfigFromPool()
		counterMutex.Unlock()
	}
	
	waitGroup.Done()
}



func main() {
	os.Remove(filepath.Join(LogDir, LogFile))
	switchToRandomConfigFromPool()
	
	timeStart := time.Now()

	counterMutex = new(sync.Mutex)
	waitGroup = new(sync.WaitGroup)
	
	waitGroup.Add(goroutinesCount)
	
	for i := 0; i < goroutinesCount; i++ {
		go logRoutine(i)
	}
	
	waitGroup.Wait()
	log.Flush()
	
	timeEnd := time.Now()
	duration := timeEnd.Sub(timeStart)
	averageLoggerReplaceFrequency := float32(loggerReplacements) / (float32(duration.Nanoseconds()) / 1e9)

	gotCount, err := test.CountSequencedRowsInFile(filepath.Join(LogDir, LogFile))
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("Logger replaced %d times. Average replacement frequency: %f times / second. Output log is consistent: no log messages are missing or come in incorrect order.\n", loggerReplacements, averageLoggerReplaceFrequency)
	
	if counter == gotCount {
		fmt.Println("PASS! Output is valid")
	} else {
		fmt.Println("ERROR! Output not valid")
	}
}
