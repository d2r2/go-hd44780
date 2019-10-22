package hd44780

import logger "github.com/d2r2/go-logger"

// You can manage verbosity of log output
// in the package by changing last parameter value.
var lg = logger.NewPackageLogger("hd44780",
	logger.DebugLevel,
	// logger.InfoLevel,
)
