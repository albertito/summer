package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
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

type Progress struct {
	start time.Time

	wg sync.WaitGroup
	mu sync.Mutex

	matched, modified, missing, corrupted int64

	done chan bool
}

func NewProgress() *Progress {
	p := &Progress{
		start: time.Now(),
		done:  make(chan bool),
	}
	p.wg.Add(1)
	go p.periodicPrint()
	return p
}

func (p *Progress) Stop() {
	p.done <- true
	p.wg.Wait()
}

func (p *Progress) periodicPrint() {
	defer p.wg.Done()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	if *quiet {
		<-p.done
		return
	}

	for {
		select {
		case <-p.done:
			p.print()
			if !*verbose {
				fmt.Printf("\n")
			}
			return
		case <-ticker.C:
			p.print()
		}
	}
}

func (p *Progress) print() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Usually we just overwrite the previous line.
	// But when verbose, just print them.
	prefix := "\r"
	suffix := ""
	if *verbose {
		prefix = ""
		suffix = "\n"
	}

	fmt.Printf(
		prefix+"%v: %d matched, %d modified, %d new, %d corrupted"+suffix,
		time.Since(p.start).Round(time.Second),
		p.matched, p.modified, p.missing, p.corrupted,
	)
}

func (p *Progress) PrintCorrupted(path string, expected, got ChecksumV1) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.corrupted++
	Printf("%q: FILE CORRUPTED - expected:%x, got:%x",
		path, expected.CRC32C, got.CRC32C)
}

func (p *Progress) PrintNew(path string, cs ChecksumV1) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.missing++
	Verbosef("%q: writing checksum (checksum:%x, mtime:%d)",
		path, cs.CRC32C, cs.ModTimeUsec)
}

func (p *Progress) PrintMissing(path string, cs *ChecksumV1) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.missing++
	if cs == nil {
		Verbosef("%q: missing checksum attribute", path)
	} else {
		Verbosef("%q: missing checksum attribute, adding it "+
			"(checksum:%x, mtime:%d)",
			path, cs.CRC32C, cs.ModTimeUsec)
	}
}

func (p *Progress) PrintModified(path string, old, new_ ChecksumV1) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.modified++
	Verbosef("%q: file modified (not corrupted) "+
		"(checksum: %x -> %x, mtime: %d -> %d)",
		path, old.CRC32C, new_.CRC32C, old.ModTimeUsec, new_.ModTimeUsec)
}

func (p *Progress) PrintMatched(path string, cs ChecksumV1) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.matched++
	Verbosef("%q: match (checksum:%x, mtime:%d)",
		path, cs.CRC32C, cs.ModTimeUsec)
}
