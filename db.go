package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"flag"
	"os"
	"path/filepath"

	"github.com/pkg/xattr"

	_ "github.com/mattn/go-sqlite3"
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

type SqliteDB struct {
	root string
	db   *sql.DB
}

const createTableV1 = `
	create table if not exists checksums (
		path string primary key,
		crc32c integer,
		modtimeusec integer
	);
`

func OpenSqliteDB(dbPath, root string) (*SqliteDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(createTableV1); err != nil {
		return nil, err
	}

	return &SqliteDB{root, db}, nil
}

func (s *SqliteDB) Has(f *os.File) (bool, error) {
	path, err := filepath.Rel(s.root, f.Name())
	if err != nil {
		return false, err
	}

	q := 0
	err = s.db.QueryRow(
		"select count(1) from checksums where path = ?",
		path).Scan(&q)
	return q == 1, err
}

func (s *SqliteDB) Read(f *os.File) (ChecksumV1, error) {
	cs := ChecksumV1{}
	path, err := filepath.Rel(s.root, f.Name())
	if err != nil {
		return cs, err
	}

	err = s.db.QueryRow(
		"select crc32c, modtimeusec from checksums where path = ?",
		path).Scan(&cs.CRC32C, &cs.ModTimeUsec)
	return cs, err
}

func (s *SqliteDB) Write(f *os.File, cs ChecksumV1) error {
	if *dryRun {
		return nil
	}

	path, err := filepath.Rel(s.root, f.Name())
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		"insert or replace into checksums "+
			"(path, crc32c, modtimeusec) values(?, ?, ?)",
		path, cs.CRC32C, cs.ModTimeUsec)
	return err
}

func (s *SqliteDB) Close() error {
	return s.db.Close()
}
