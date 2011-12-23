// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	log "github.com/cihub/sealog"
	library "../testlibrary"
	"fmt"
)

func loadAppConfig() {
	appConfig := `
<sealog type="sync">
    <outputs formatid="app">
        <console />
    </outputs>
    <formats>
        <format id="app" format="app: [%LEV] %Msg%n" />
    </formats>
</sealog>
`
	logger, err := log.LoggerFromBytes([]byte(appConfig))
	if err != nil {
		fmt.Println(err)
		return
	}
	log.ReplaceLogger(logger)
}

func calcF() {
	x := 1
	y := 2
	log.Debug("Calculating F")
	result := library.CalculateF(x,y)
	log.Debug("Got F = %d", result)
}

func sameOutputConfig() {
	libConfig := `
<sealog type="sync">
    <outputs formatid="library">
        <console />
    </outputs>
    <formats>
        <format id="library" format="testlibrary: [%LEV] %Msg%n" />
    </formats>
</sealog>
`
	logger, err := log.LoggerFromBytes([]byte(libConfig))
	if err != nil {
		fmt.Println(err)
		return
	}
	library.UseLogger(logger)

	calcF()
}

func specialOutputConfig() {
	libConfig := `
<sealog type="sync">
    <outputs formatid="library">
        <console />
    </outputs>
    <formats>
        <format id="library" format="library + app: %Msg [%LEV] %n" />
    </formats>
</sealog>
`
	logger, err := log.LoggerFromBytes([]byte(libConfig))
	if err != nil {
		fmt.Println(err)
		return
	}
	library.UseLogger(logger)

	calcF()
}

func main() {
	defer library.FlushLog()
	defer log.Flush()
	log.Info("App started")
	loadAppConfig()	
	log.Info("Config loaded")
	sameOutputConfig()
	log.Info("Same output config tested")
	specialOutputConfig()
	log.Info("Special output config tested")
	log.Info("App finished")
	
	
}