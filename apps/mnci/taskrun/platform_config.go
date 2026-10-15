package taskrun

import "runtime"

var isWindows = runtime.GOOS == "windows"

// Platform is part of every cache key: a binary built on one OS is not the
// output of the same task on another.
var Platform = runtime.GOOS + "/" + runtime.GOARCH
