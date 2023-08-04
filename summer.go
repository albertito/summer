package main

import (
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

const usage = `# summer 🌞 🏖

Utility to detect accidental data corruption (e.g. bitrot, storage media
problems).  Not intended to detect malicious modification.

Checksums are written to/read from each files' extended attributes by default,
or to a separate database file (with the -db flag).

Usage:

  summer update <dir>
      Verify checksums in the given directory, and update them for new or
      changed files.
  summer verify <dir>
      Verify checksums in the given directory.
  summer generate <dir>
      Write checksums for the given directory. Files with pre-existing
      checksums are left untouched, and checksums are not verified.
      Useful when generating checksums for a lot of files for the first time,
      as is faster to resume work if interrupted.
  summer version
      Print software version information.

Flags:
`

var (
	dbPath        = flag.String("db", "", "database to read from/write to")
	oneFilesystem = flag.Bool("x", false, "don't cross filesystem boundaries")
)

var options = struct {
	// Database to use.
	db DB

	// Do not cross filesystem boundaries.
	oneFilesystem bool
}{}

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), usage)
	flag.PrintDefaults()
}

func main() {
	var err error

	flag.Usage = Usage
	flag.Parse()

	options.oneFilesystem = *oneFilesystem

	op := flag.Arg(0)
	root := flag.Arg(1)

	if op != "version" && root == "" {
		Usage()
		os.Exit(1)
	}

	options.db = XattrDB{}
	if *dbPath != "" {
		options.db, err = OpenSqliteDB(*dbPath, root)
		if err != nil {
			Fatalf("%q: %v", *dbPath, err)
		}
	}
	defer options.db.Close()

	switch op {
	case "generate":
		err = generate(root)
	case "verify":
		err = verify(root)
	case "update":
		err = update(root)
	case "version":
		PrintVersion()
	default:
		Fatalf("unknown command %q", op)
	}

	if err != nil {
		Fatalf("%v", err)
	}
}

var crc32c = crc32.MakeTable(crc32.Castagnoli)

type ChecksumV1 struct {
	// CRC32C of the file contents.
	CRC32C uint32

	// Modification time of the file when the checksum was computed.
	// In Unix microseconds.
	ModTimeUsec int64
}

func openAndInfo(path string, d fs.DirEntry, err error, rootDev uint64) (bool, *os.File, fs.FileInfo, error) {
	if err != nil {
		return false, nil, nil, err
	}
	if d.IsDir() || !d.Type().IsRegular() {
		return false, nil, nil, nil
	}

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

func getDevice(info fs.FileInfo) uint64 {
	return info.Sys().(*syscall.Stat_t).Dev
}

func getDeviceForPath(path string) uint64 {
	fi, err := os.Stat(path)
	if err != nil {
		// Doesn't matter, because we'll get an error during WalkDir.
		return 0
	}
	return getDevice(fi)
}

func generate(root string) error {
	rootDev := getDeviceForPath(root)
	p := NewProgress()
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

	err := filepath.WalkDir(root, fn)
	return err
}

func verify(root string) error {
	rootDev := getDeviceForPath(root)
	p := NewProgress()
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

	err := filepath.WalkDir(root, fn)

	if p.corrupted > 0 && err == nil {
		err = fmt.Errorf("detected %d corrupted files", p.corrupted)
	}
	return err
}

func update(root string) error {
	rootDev := getDeviceForPath(root)
	p := NewProgress()
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

	err := filepath.WalkDir(root, fn)

	if p.corrupted > 0 && err == nil {
		err = fmt.Errorf("detected %d corrupted files", p.corrupted)
	}
	return err
}
