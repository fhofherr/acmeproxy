package certutil

import "flag"

// FlagUpdate signals that the tests should update their golden files.
// Note that not all golden files used throughout the package end in .golden.
// Especially key files have a different extension which is more practical.
var FlagUpdate = flag.Bool("update", false, "update golden files")
