
Basic tests but storing in a sqlite3 database.
This is enough to exercise the backend.

  $ alias summer="$TESTDIR/../summer"
  $ touch empty
  $ echo marola > hola

  $ summer -db=db.sqlite3 generate .
  \r (no-eol) (esc)
  0s: 0 matched, 0 modified, 3 new, 0 corrupted
  $ summer -db=db.sqlite3 verify .
  \r (no-eol) (esc)
  0s: 2 matched, 1 modified, 0 new, 0 corrupted
  $ summer -db=db.sqlite3 update .
  \r (no-eol) (esc)
  0s: 2 matched, 1 modified, 0 new, 0 corrupted

Check that the root path doesn't confuse us.

  $ summer -db=db.sqlite3 -v verify $PWD
  ".*/db.sqlite3": file modified \(not corrupted\) \(checksum: \w+ -> \w+, mtime: \d+ -> \d+\) (re)
  ".*/empty": match \(checksum:0, mtime:\d+\) (re)
  ".*/hola": match \(checksum:\w+, mtime:\d+\) (re)
  0s: 2 matched, 1 modified, 0 new, 0 corrupted

Force a write error to check it is appropriately handled.

  $ summer "-db=file:db.sqlite3?mode=ro" generate .
  \r (no-eol) (esc)
  0s: 0 matched, 0 modified, 0 new, 0 corrupted
  attempt to write a readonly database
  [1]

Check errors when we cannot open the database file.

  $ summer -db=/proc/doesnotexist verify .
  "/proc/doesnotexist": unable to open database file: no such file or directory
  [1]

  $ summer -db=/dev/null verify .
  "/dev/null": attempt to write a readonly database
  [1]

