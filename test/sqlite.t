
Basic tests but storing in a sqlite3 database.
This is enough to exercise the backend.

  $ alias summer="$TESTDIR/../summer"
  $ touch empty
  $ echo marola > hola

  $ summer -db=db.sqlite3 generate .
  3 checksums written
  $ summer -db=db.sqlite3 verify .
  2 matched, 1 modified, 0 new, 0 corrupted
  $ summer -db=db.sqlite3 update .
  2 matched, 1 modified, 0 new, 0 corrupted

Check that the root path doesn't confuse us.

  $ summer -db=db.sqlite3 -v verify $PWD
  ".*/db.sqlite3": file modified \(not corrupted\), updating (re)
  ".*/empty": match (re)
  ".*/hola": match (re)
  2 matched, 1 modified, 0 new, 0 corrupted

Force a write error to check it is appropriately handled.

  $ summer "-db=file:db.sqlite3?mode=ro" generate .
  . checksums written (re)
  attempt to write a readonly database
  [1]

Check errors when we cannot open the database file.

  $ summer -db=/proc/doesnotexist verify .
  "/proc/doesnotexist": unable to open database file: no such file or directory
  [1]

  $ summer -db=/dev/null verify .
  "/dev/null": attempt to write a readonly database
  [1]

