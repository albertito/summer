
Tests for how to handle file access issues.
Note we put our test paths in "root/" so the database doesn't
interfere.

  $ alias summer="$TESTDIR/../summer"
  $ mkdir root
  $ touch root/empty
  $ echo marola > root/hola

  $ summer -db=db.sqlite3 generate root/
  \r (no-eol) (esc)
  0s: 0 matched, 0 modified, 2 new, 0 corrupted

  $ summer -db=db.sqlite3 verify root/
  \r (no-eol) (esc)
  0s: 2 matched, 0 modified, 0 new, 0 corrupted
  $ chmod 0000 root/empty
  $ summer -db=db.sqlite3 verify root/
  \r (no-eol) (esc)
  0s: 0 matched, 0 modified, 0 new, 0 corrupted
  open root/empty: permission denied
  [1]
