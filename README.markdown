Sealog
=======

Sealog is a powerful and easy-to-learn logging framework that provides functionality for flexible dispatching, filtering, and formatting.
Natively written in the [Go](http://golang.org/) programming language. 

Features
------------------

* Xml configuring to be able to change logger parameters without recompilation
* Changing configurations on the fly without app restart
* Possibility to set different log configurations for different project files and functions
* Adjustable message formatting
* Simultaneous log output to multiple streams
* Choosing logger priority strategy to minimize performance impact
* Different output writers
  * Console writer
  * File writer 
  * Buffered file writer (Chunk writer)
  * Rolling log writer (Logging with rotation)
  * SMTP writer
  * TCP/UDP network writers
  * others... (See [Wiki](https://github.com/cihub/sealog/wiki))
* Log message wrappers (JSON, XML, etc.)
* Global variables and functions for easy usage in standalone apps
* Functions for flexible usage in libraries

Quick-start
-----------

```go
package main

import log "github.com/cihub/sealog"

func main() {
    defer log.Flush()
    log.Info("Hello from Sealog!")
}
```

Installation
------------

* If you don't have the Go development environment installed, visit the 
[Getting Started](http://golang.org/doc/install.html) document and follow the instructions
* goinstall github.com/cihub/sealog

Documentation
---------------

Sealog has a github wiki page that contains a detailed Sealog reference: https://github.com/cihub/sealog/wiki

Examples
---------------

Sealog examples can be found in 'sealog/examples' folder. Current list of examples: 

* **examples/exceptions** - demonstrates constraints and exceptions
* **examples/outputs** - demonstrates dispatchers and writers
* **examples/formats** - demonstrates formats
* **examples/types** - demonstrates logger types

Issues
---------------

Feel free to add any issues that could make Sealog better: https://github.com/cihub/sealog/issues

