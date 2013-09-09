Seelog
=======

Seelog is a powerful and easy-to-learn logging framework that provides functionality for flexible dispatching, filtering, and formatting log messages.
It is natively written in the [Go](http://golang.org/) programming language. 

[![Build Status](https://drone.io/github.com/cihub/seelog/status.png)](https://drone.io/github.com/cihub/seelog/latest)

Features
------------------

* Xml configuring to be able to change logger parameters without recompilation
* Changing configurations on the fly without app restart
* Possibility to set different log configurations for different project files and functions
* Adjustable message formatting
* Simultaneous log output to multiple streams
* Choosing logger priority strategy to minimize performance hit
* Different output writers
  * Console writer
  * File writer 
  * Buffered writer (Chunk writer)
  * Rolling log writer (Logging with rotation)
  * SMTP writer
  * Others... (See [Wiki](https://github.com/cihub/seelog/wiki))
* Log message wrappers (JSON, XML, etc.)
* Global variables and functions for easy usage in standalone apps
* Functions for flexible usage in libraries

Quick-start
-----------

```go
package main

import log "github.com/cihub/seelog"

func main() {
    defer log.Flush()
    log.Info("Hello from Seelog!")
}
```

Installation
------------

If you don't have the Go development environment installed, visit the 
[Getting Started](http://golang.org/doc/install.html) document and follow the instructions. Once you're ready, execute the following command:

```
go get -u github.com/cihub/seelog
```

Documentation
---------------

Seelog has github wiki pages, which contain detailed how-tos references: https://github.com/cihub/seelog/wiki

Examples
---------------

Seelog examples can be found here: [seelog-examples](https://github.com/cihub/seelog-examples)

Issues
---------------

Feel free to push issues that could make Seelog better: https://github.com/cihub/seelog/issues

