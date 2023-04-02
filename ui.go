package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

var (
	verbose = flag.Bool("v", false, "verbose mode (list each file)")
	quiet   = flag.Bool("q", false, "quiet mode")
)

func Verbosef(format string, args ...interface{}) {
	if *verbose {
		fmt.Printf(format+"\n", args...)
	}
}

func Printf(format string, args ...interface{}) {
	if !*quiet {
		fmt.Printf(format+"\n", args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}

func PrintWritten(written int64) {
	Printf("%d checksums written", written)
}

func PrintSummary(matched, modified, missing, corrupted int64) {
	Printf("%d matched, %d modified, %d new, %d corrupted",
		matched, modified, missing, corrupted)
}

func PrintCorrupted(path string, expected, got ChecksumV1) {
	Printf("%q: FILE CORRUPTED - expected:%x, got:%x",
		path, expected.CRC32C, got.CRC32C)
}

func PrintMissing(path string) {
	Verbosef("%q: missing checksum attribute, adding it", path)
}

func PrintModified(path string) {
	Verbosef("%q: file modified (not corrupted), updating", path)
}

func PrintMatched(path string) {
	Verbosef("%q: match", path)
}

func PrintVersion() {
	info, _ := debug.ReadBuildInfo()
	rev := ""
	ts := time.Time{}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.time":
			ts, _ = time.Parse(time.RFC3339, s.Value)
		}
	}
	Printf("summer version %s (%s)", rev, ts)
}
