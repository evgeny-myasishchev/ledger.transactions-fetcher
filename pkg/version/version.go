package version

// Version is the app-global version string, which should be substituted with a
// real value during build
var Version = "UNKNOWN"

// AppName is a name of a service
// should be in sync with Makefile
var AppName = "transactions-fetcher"

// GitHash injected build time (see Makefile)
var GitHash = "TBD"

// GitRef injected build time (see Makefile)
var GitRef = "TBD"

// GitURL injected build time (see Makefile)
var GitURL = "TBD"
