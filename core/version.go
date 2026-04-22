package main

import (
	"flag"
	"fmt"
	"os"
)

// injected by goreleaser on build
var (
	version     = "dev"
	commit      = "none"
	date        = "unknown"
	showVersion bool
)

func showVersionExit() {
	fmt.Printf("airstats %s (commit %s) build %s\n", version, commit, date)
	os.Exit(0)
}

func init() {
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showVersion, "v", false, "display version")
}
