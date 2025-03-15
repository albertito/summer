package main

import (
	"errors"
	"io/fs"
	"os"
	"syscall"
	"testing"
)

// Tests for errors that are not feasible to cover by the end to end tests.
// For example, db.Read() calls are usually preceded by a db.Has(), so we
// can't easily simulate Read seeing an "xattr not found" in an end-to-end
// test.

func init() {
	// Initialize the subset options, since they are used as part of the walk.
	options.subset, _ = NewSubset()
}

func TestDBReadError(t *testing.T) {
	f, err := os.Open("/dev/null")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	db := XattrDB{}
	_, err = db.Read(f)

	if !errors.Is(err, syscall.ENODATA) {
		t.Fatalf("expected ENODATA, got %v", err)
	}
}

var testErr = errors.New("test error")

type fakeDirEntry struct{}

func (f fakeDirEntry) Name() string {
	return "fake"
}

func (f fakeDirEntry) IsDir() bool {
	return false
}

func (f fakeDirEntry) Type() fs.FileMode {
	return fs.FileMode(0777)
}

func (f fakeDirEntry) Info() (os.FileInfo, error) {
	return nil, testErr
}

func TestOpenAndInfoError(t *testing.T) {
	ok, _, _, err := openAndInfo("fake", fakeDirEntry{}, nil, 0)
	if !ok || err != testErr {
		t.Fatalf("expected ok, testErr, got %v, %v", ok, err)
	}
}

type fakeDB struct {
	hasAttr bool
	hasErr  error

	readChecksum ChecksumV1
	readErr      error

	writeErr error
}

func (db fakeDB) Has(f *os.File) (bool, error) {
	return db.hasAttr, db.hasErr
}

func (db fakeDB) Read(f *os.File) (ChecksumV1, error) {
	return db.readChecksum, db.readErr
}

func (db fakeDB) Write(f *os.File, c ChecksumV1) error {
	return db.writeErr
}

func (db fakeDB) Close() error {
	return nil
}

func TestWalkingFunctionsHandleDBErrors(t *testing.T) {
	// Test how the various walking functions (generate, update, verify)
	// handle errors from the database.
	cases := []struct {
		fn       walkFn
		db       fakeDB
		expected error
	}{
		{generate, fakeDB{}, nil},
		{generate, fakeDB{hasErr: testErr}, testErr},
		{generate, fakeDB{writeErr: testErr}, testErr},

		{verify, fakeDB{}, nil},
		{verify, fakeDB{hasErr: testErr}, testErr},

		{update, fakeDB{}, nil},
		{update, fakeDB{hasAttr: true}, nil},
		{update, fakeDB{hasErr: testErr}, testErr},
		{update, fakeDB{hasAttr: true, readErr: testErr}, testErr},
	}

	p := NewProgress(false)

	for _, c := range cases {
		f, err := os.Open("/dev/null")
		if err != nil {
			t.Fatal(err)
		}
		info, err := f.Stat()
		if err != nil {
			t.Fatal(err)
		}

		options.db = c.db

		err = c.fn(f, info, p)
		if err != c.expected {
			t.Fatalf("expected %v, got %v", c.expected, err)
		}
	}
}
