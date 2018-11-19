package version

var (
	// Version is overridden by go build flags
	Version = "???"
	// BuildDate is overridden by go build flags
	BuildDate = "???"
	// Package is the package name used to build this project, overridden by go build flags
	Package = "???"
	// Branch is which branch this build was created on
	Branch = "???"
	// Commit is which git commit this build was created at
	Commit = "???"
)
