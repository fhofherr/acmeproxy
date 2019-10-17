package version

var (
	// BuildTime contains the time the currently executing acmeproxy binary
	// was build.
	BuildTime string
	// GitHash references the full git commit hash the currently executing
	// acmeproxy binary was build from.
	GitHash string
	// GitTag references the git tag the currently executing acmeproxy binary
	// was build from. GitTag may be empty if the binary was not build from a
	// tag.
	GitTag string
)
