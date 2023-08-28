package main

import (
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/term"
)

const usage = `# summer 🌞 🏖

Utility to detect accidental data corruption (e.g. bitrot, storage media
problems).  Not intended to detect malicious modification.

Checksums are written to/read from each file's extended attributes.

Paths given can be files or directories. If a directory is given, it is
processed recursively.

Usage:

  summer [flags] update <paths>
      Verify checksums in the given paths, and update them for new or changed
      files.
  summer [flags] verify <paths>
      Verify checksums in the given paths.
  summer [flags] generate <paths>
      Write checksums for the given paths. Files with pre-existing checksums
      are left untouched, and checksums are not verified.
      Useful when generating checksums for a lot of files for the first time,
      as is faster to resume work if interrupted.
  summer [flags] version
      Print software version information.

Flags:
`

// Flags.
var (
	oneFilesystem = flag.Bool("x", false, "don't cross filesystem boundaries")
	forceTTY      = flag.Bool("forcetty", false, "force TTY output")
	exclude       = &RepeatedStringFlag{}
	excludeRe     = &RepeatedStringFlag{}
)

var options = struct {
	// Database to use.
	db DB

	// Do not cross filesystem boundaries.
	oneFilesystem bool

	// Whether output is a TTY.
	isTTY bool

	// Paths to exclude.
	exclude map[string]bool

	// Regexp patterns to exclude.
	excludeRe []*regexp.Regexp
}{}

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), usage)
	flag.PrintDefaults()
}

func main() {
	var err error

	flag.Var(exclude, "exclude",
		"exclude these paths (can be repeated)")
	flag.Var(excludeRe, "excludere",
		"exclude paths matching this regexp (can be repeated)")

	flag.Usage = Usage
	flag.Parse()

	options.oneFilesystem = *oneFilesystem
	options.isTTY = *forceTTY || term.IsTerminal(int(os.Stdout.Fd()))

	options.exclude = map[string]bool{}
	for _, s := range *exclude {
		options.exclude[filepath.Clean(s)] = true
	}

	for _, s := range *excludeRe {
		options.excludeRe = append(options.excludeRe, regexp.MustCompile(s))
	}

	op := flag.Arg(0)
	roots := []string{}
	if flag.NArg() > 1 {
		roots = flag.Args()[1:]
	}

	if op != "version" && len(roots) == 0 {
		Usage()
		os.Exit(1)
	}

	options.db = XattrDB{}
	defer options.db.Close()

	switch op {
	case "generate":
		err = walk(roots, generate)
	case "verify":
		err = walk(roots, verify)
	case "update":
		err = walk(roots, update)
	case "version":
		PrintVersion()
	default:
		Fatalf("unknown command %q", op)
	}

	if err != nil {
		Fatalf("%v", err)
	}
}

func isExcluded(path string) bool {
	if options.exclude[path] {
		return true
	}
	for _, re := range options.excludeRe {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

var crc32c = crc32.MakeTable(crc32.Castagnoli)

type ChecksumV1 struct {
	// CRC32C of the file contents.
	CRC32C uint32

	// Modification time of the file when the checksum was computed.
	// In Unix microseconds.
	//
	// Because we do not lock the file, it could be modified as we read it,
	// and then depending on the order of operations, the checksum could be
	// wrong and cause a false positive.
	// To avoid this, we always read the modification time prior to reading
	// the file contents. That way, if the file changes while we are reading
	// it, it should be detected later as modified instead of corrupted.
	//
	// This relies on mtime having enough resolution to detect the change. On
	// some filesystems that may not be the case, and a file modified very
	// quickly may not be detected as modified. This is a limitation of the
	// filesystem, and there is nothing we can do about it.
	ModTimeUsec int64
}

func generate(fd *os.File, info fs.FileInfo, p *Progress) error {
	hasAttr, err := options.db.Has(fd)
	if err != nil {
		return err
	}
	if hasAttr {
		// Skip files that already have a checksum.
		return nil
	}

	h := crc32.New(crc32c)
	_, err = io.Copy(h, fd)
	if err != nil {
		return err
	}

	csum := ChecksumV1{
		CRC32C:      h.Sum32(),
		ModTimeUsec: info.ModTime().UnixMicro(),
	}

	err = options.db.Write(fd, csum)
	if err != nil {
		return err
	}

	p.PrintNew(fd.Name(), csum)
	return nil
}

func verify(fd *os.File, info fs.FileInfo, p *Progress) error {
	hasAttr, err := options.db.Has(fd)
	if err != nil {
		return err
	}
	if !hasAttr {
		p.PrintMissing(fd.Name(), nil)
		return nil
	}

	csumFromFile, err := options.db.Read(fd)
	if err != nil {
		return err
	}

	h := crc32.New(crc32c)
	_, err = io.Copy(h, fd)
	if err != nil {
		return err
	}

	csumComputed := ChecksumV1{
		CRC32C:      h.Sum32(),
		ModTimeUsec: info.ModTime().UnixMicro(),
	}

	if csumFromFile.ModTimeUsec != csumComputed.ModTimeUsec {
		p.PrintModified(fd.Name(), csumFromFile, csumComputed)
	} else if csumFromFile.CRC32C != csumComputed.CRC32C {
		p.PrintCorrupted(fd.Name(), csumFromFile, csumComputed)
	} else {
		p.PrintMatched(fd.Name(), csumComputed)
	}

	return nil
}

func update(fd *os.File, info fs.FileInfo, p *Progress) error {
	// Compute checksum from the current state.
	h := crc32.New(crc32c)
	_, err := io.Copy(h, fd)
	if err != nil {
		return err
	}

	csumComputed := ChecksumV1{
		CRC32C:      h.Sum32(),
		ModTimeUsec: info.ModTime().UnixMicro(),
	}

	// Read the saved checksum (if any).
	hasAttr, err := options.db.Has(fd)
	if err != nil {
		return err
	}
	if !hasAttr {
		// Attribute is missing. Expected for newly created files.
		p.PrintMissing(fd.Name(), &csumComputed)
		return options.db.Write(fd, csumComputed)
	}

	csumFromFile, err := options.db.Read(fd)
	if err != nil {
		return err
	}

	if csumFromFile.ModTimeUsec != csumComputed.ModTimeUsec {
		// File modified. Expected for updated files.
		p.PrintModified(fd.Name(), csumFromFile, csumComputed)
		return options.db.Write(fd, csumComputed)
	} else if csumFromFile.CRC32C != csumComputed.CRC32C {
		p.PrintCorrupted(fd.Name(), csumFromFile, csumComputed)
	} else {
		p.PrintMatched(fd.Name(), csumComputed)
	}

	return nil
}
