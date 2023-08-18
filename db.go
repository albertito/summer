package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"os"

	"github.com/pkg/xattr"
)

var dryRun = flag.Bool("n", false,
	"dry-run mode (do not write anything)")

type DB interface {
	Has(f *os.File) (bool, error)
	Read(f *os.File) (ChecksumV1, error)
	Write(f *os.File, cs ChecksumV1) error
	Close() error
}

type XattrDB struct{}

func (_ XattrDB) Has(f *os.File) (bool, error) {
	attrs, err := xattr.FList(f)
	for _, a := range attrs {
		if a == "user.summer-v1" {
			return true, err
		}
	}
	return false, err
}

func (_ XattrDB) Read(f *os.File) (ChecksumV1, error) {
	val, err := xattr.FGet(f, "user.summer-v1")
	if err != nil {
		return ChecksumV1{}, err
	}

	buf := bytes.NewReader(val)
	c := ChecksumV1{}
	err = binary.Read(buf, binary.LittleEndian, &c)
	return c, err
}

func (_ XattrDB) Write(f *os.File, cs ChecksumV1) error {
	if *dryRun {
		return nil
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, cs)
	if err != nil {
		// We control the struct, it should never panic.
		panic(err)
	}
	return xattr.FSet(f, "user.summer-v1", buf.Bytes())
}

func (_ XattrDB) Close() error {
	return nil
}
