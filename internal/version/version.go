package version

var version = "dev" // overridden at build time

func String() string { return version }
