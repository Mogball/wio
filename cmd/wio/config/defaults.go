package config

import "wio/cmd/wio/constants"

var IdeDefault = "none"
var PackageAllPlatform = "native"
var PackageAllFramework = "c++"
var PackageAllBoard = "null"
var DefaultTargetDefaults = "default"
var PortDefaults = "9600"
var BaudRateDefaults = 9600

var FrameworkDefaults = map[string]string{
    "atmelavr": "cosa",
    "native":   "c++",
}

var BoardDefaults = map[string]string{
    "atmelavr": "uno",
}

var TargetSourceDefault = map[constants.Type]string{
    constants.APP: "src",
    constants.PKG: "tests",
}
