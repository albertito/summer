package main

import (
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
      Write checksums for the given directory. Pre-existing checksums are
      overwritten without verification.
  summer version
      Print software version information.

Flags:
`

var (
	dbPath = flag.String("db", "", "database to read from/write to")
)

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), usage)
	flag.PrintDefaults()
}

func main() {
	var err error

	flag.Usage = Usage
	flag.Parse()

	op := flag.Arg(0)
	root := flag.Arg(1)

	if op != "version" && root == "" {
		Usage()
		os.Exit(1)
	}

	var db DB = XattrDB{}
	if *dbPath != "" {
		db, err = OpenSqliteDB(*dbPath, root)
		if err != nil {
			Fatalf("%q: %v", *dbPath, err)
		}
	}
	defer db.Close()

	switch op {
	case "generate":
		err = generate(db, root)
	case "verify":
		err = verify(db, root)
	case "update":
		err = update(db, root)
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

func isFileRelevant(path string, d fs.DirEntry, err error) bool {
	if err != nil {
		return false
	}
	if d.IsDir() {
		return false
	}
	return d.Type().IsRegular()
}

func openAndInfo(path string, d fs.DirEntry) (*os.File, fs.FileInfo, error) {
	info, err := d.Info()
	if err != nil {
		return nil, nil, err
	}
	fd, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	return fd, info, nil
}

func generate(db DB, root string) error {
	p := NewProgress()
	defer p.Stop()
	fn := func(path string, d fs.DirEntry, err error) error {
		if !isFileRelevant(path, d, err) {
			return err
		}

		fd, info, err := openAndInfo(path, d)
		if err != nil {
			return err
		}
		defer fd.Close()

		h := crc32.New(crc32c)
		_, err = io.Copy(h, fd)
		if err != nil {
			return err
		}

		csum := ChecksumV1{
			CRC32C:      h.Sum32(),
			ModTimeUsec: info.ModTime().UnixMicro(),
		}

		err = db.Write(fd, csum)
		if err != nil {
			return err
		}

		p.PrintNew(path)
		return nil
	}

	err := filepath.WalkDir(root, fn)
	return err
}

func verify(db DB, root string) error {
	p := NewProgress()
	defer p.Stop()

	fn := func(path string, d fs.DirEntry, err error) error {
		if !isFileRelevant(path, d, err) {
			return err
		}

		fd, info, err := openAndInfo(path, d)
		if err != nil {
			return err
		}
		defer fd.Close()

		hasAttr, err := db.Has(fd)
		if err != nil {
			return err
		}
		if !hasAttr {
			p.PrintMissing(path)
			return nil
		}

		csumFromFile, err := db.Read(fd)
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
			p.PrintModified(path)
		} else if csumFromFile.CRC32C != csumComputed.CRC32C {
			p.PrintCorrupted(path, csumFromFile, csumComputed)
		} else {
			p.PrintMatched(path)
		}

		return nil
	}

	err := filepath.WalkDir(root, fn)

	if p.corrupted > 0 && err == nil {
		err = fmt.Errorf("detected %d corrupted files", p.corrupted)
	}
	return err
}

func update(db DB, root string) error {
	p := NewProgress()
	defer p.Stop()

	fn := func(path string, d fs.DirEntry, err error) error {
		if !isFileRelevant(path, d, err) {
			return err
		}

		fd, info, err := openAndInfo(path, d)
		if err != nil {
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
		hasAttr, err := db.Has(fd)
		if err != nil {
			return err
		}
		if !hasAttr {
			// Attribute is missing. Expected for newly created files.
			p.PrintMissing(path)
			return db.Write(fd, csumComputed)
		}

		csumFromFile, err := db.Read(fd)
		if err != nil {
			return err
		}

		if csumFromFile.ModTimeUsec != csumComputed.ModTimeUsec {
			// File modified. Expected for updated files.
			p.PrintModified(path)
			return db.Write(fd, csumComputed)
		} else if csumFromFile.CRC32C != csumComputed.CRC32C {
			p.PrintCorrupted(path, csumFromFile, csumComputed)
		} else {
			p.PrintMatched(path)
		}

		return nil
	}

	err := filepath.WalkDir(root, fn)

	if p.corrupted > 0 && err == nil {
		err = fmt.Errorf("detected %d corrupted files", p.corrupted)
	}
	return err
}
