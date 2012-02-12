// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	log "github.com/cihub/sealog"
	library "github.com/cihub/sealog/examples/library/library"
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
	logger, err := log.LoggerFromConfigAsBytes([]byte(appConfig))
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

// Same config for both library and app
func sameOutputConfig() {
	libConfig := `
<sealog type="sync">
    <outputs formatid="library">
        <console />
    </outputs>
    <formats>
        <format id="library" format="library + app: [%LEV] %Msg%n" />
    </formats>
</sealog>
`
	logger, err := log.LoggerFromConfigAsBytes([]byte(libConfig))
	if err != nil {
		fmt.Println(err)
		return
	}
	log.ReplaceLogger(logger)
	library.UseLogger(logger)
}

// Special config for library (app config is not changed)
func specialOutputConfig() {
	libConfig := `
<sealog type="sync">
    <outputs formatid="library">
        <console />
    </outputs>
    <formats>
        <format id="library" format="library: %Msg [%LEV] %n" />
    </formats>
</sealog>
`
	logger, err := log.LoggerFromConfigAsBytes([]byte(libConfig))
	if err != nil {
		fmt.Println(err)
		return
	}
	library.UseLogger(logger)
}

func main() {
	defer library.FlushLog()
	defer log.Flush()
	loadAppConfig()	
	log.Info("App started")
	log.Info("Config loaded")

	// Disable library log
	log.Info("* Disabled library log test")
	library.DisableLog();
	calcF();
	log.Info("* Disabled library log tested")

	// Use a special logger for library
	log.Info("* Special output test")
	specialOutputConfig()
	calcF();
	log.Info("* Special output tested")
	
	// Use the same logger for both app and library
	log.Info("* Same output test")
	sameOutputConfig()
	calcF();
	log.Info("* Same output tested")
		
	log.Info("App finished")
}
