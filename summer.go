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
	"syscall"

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
		err = generate(roots)
	case "verify":
		err = verify(roots)
	case "update":
		err = update(roots)
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

func openAndInfo(path string, d fs.DirEntry, err error, rootDev deviceID) (bool, *os.File, fs.FileInfo, error) {
	// Excluded check must come first, because it can be use to skip
	// directories that would otherwise cause errors.
	if isExcluded(path) {
		if d.IsDir() {
			return false, nil, nil, fs.SkipDir
		}
		return false, nil, nil, nil
	}

	if err != nil {
		return false, nil, nil, err
	}
	if d.IsDir() || !d.Type().IsRegular() {
		return false, nil, nil, nil
	}

	// It is important that we obtain fs.FileInfo at this point, before
	// reading any of the file contents, because the file could be modified
	// while we do so. See the comment on ChecksumV1.ModTimeUsec for more
	// details.
	info, err := d.Info()
	if err != nil {
		return true, nil, nil, err
	}

	fd, err := os.Open(path)
	if err != nil {
		return true, nil, nil, err
	}

	if options.oneFilesystem && rootDev != getDevice(info) {
		fd.Close()
		return false, nil, nil, fs.SkipDir
	}

	return true, fd, info, nil
}

type deviceID uint64

func getDevice(info fs.FileInfo) deviceID {
	return deviceID(info.Sys().(*syscall.Stat_t).Dev)
}

func getDeviceForPath(path string) deviceID {
	fi, err := os.Stat(path)
	if err != nil {
		// Doesn't matter, because we'll get an error during WalkDir.
		return 0
	}
	return getDevice(fi)
}

func generate(roots []string) error {
	rootDev := deviceID(0)
	p := NewProgress(options.isTTY)
	defer p.Stop()

	fn := func(path string, d fs.DirEntry, err error) error {
		ok, fd, info, err := openAndInfo(path, d, err, rootDev)
		if !ok || err != nil {
			return err
		}
		defer fd.Close()

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

		p.PrintNew(path, csum)
		return nil
	}

	for _, root := range roots {
		rootDev = getDeviceForPath(root)
		err := filepath.WalkDir(root, fn)
		if err != nil {
			return err
		}
	}
	return nil
}

func verify(roots []string) error {
	rootDev := deviceID(0)
	p := NewProgress(options.isTTY)
	defer p.Stop()

	fn := func(path string, d fs.DirEntry, err error) error {
		ok, fd, info, err := openAndInfo(path, d, err, rootDev)
		if !ok || err != nil {
			return err
		}
		defer fd.Close()

		hasAttr, err := options.db.Has(fd)
		if err != nil {
			return err
		}
		if !hasAttr {
			p.PrintMissing(path, nil)
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
			p.PrintModified(path, csumFromFile, csumComputed)
		} else if csumFromFile.CRC32C != csumComputed.CRC32C {
			p.PrintCorrupted(path, csumFromFile, csumComputed)
		} else {
			p.PrintMatched(path, csumComputed)
		}

		return nil
	}

	var err error
	for _, root := range roots {
		rootDev = getDeviceForPath(root)
		err = filepath.WalkDir(root, fn)
		if err != nil {
			break
		}
	}

	if p.corrupted > 0 && err == nil {
		err = fmt.Errorf("detected %d corrupted files", p.corrupted)
	}
	return err
}

func update(roots []string) error {
	rootDev := deviceID(0)
	p := NewProgress(options.isTTY)
	defer p.Stop()

	fn := func(path string, d fs.DirEntry, err error) error {
		ok, fd, info, err := openAndInfo(path, d, err, rootDev)
		if !ok || err != nil {
			return err
		}
		defer fd.Close()

		// Compute checksum from the current state.
		h := crc32.New(crc32c)
		_, err = io.Copy(h, fd)
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
			p.PrintMissing(path, &csumComputed)
			return options.db.Write(fd, csumComputed)
		}

		csumFromFile, err := options.db.Read(fd)
		if err != nil {
			return err
		}

		if csumFromFile.ModTimeUsec != csumComputed.ModTimeUsec {
			// File modified. Expected for updated files.
			p.PrintModified(path, csumFromFile, csumComputed)
			return options.db.Write(fd, csumComputed)
		} else if csumFromFile.CRC32C != csumComputed.CRC32C {
			p.PrintCorrupted(path, csumFromFile, csumComputed)
		} else {
			p.PrintMatched(path, csumComputed)
		}

		return nil
	}

	var err error
	for _, root := range roots {
		rootDev = getDeviceForPath(root)
		err = filepath.WalkDir(root, fn)
		if err != nil {
			break
		}
	}

	if p.corrupted > 0 && err == nil {
		err = fmt.Errorf("detected %d corrupted files", p.corrupted)
	}
	return err
}
