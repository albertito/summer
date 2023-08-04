Tests for dry-run mode when using -db. Similar to the basic ones.

  $ alias summer="$TESTDIR/../summer"

Generate test data.

  $ touch empty
  $ echo marola > hola

Generate and verify.

  $ summer -n -db=db.sqlite3 generate .
  0s: 0 matched, 0 modified, 3 new, 0 corrupted

  $ summer -n -db=db.sqlite3 verify .
  0s: 0 matched, 0 modified, 3 new, 0 corrupted

  $ summer -n -db=db.sqlite3 verify .
  0s: 0 matched, 0 modified, 3 new, 0 corrupted

Now write data for real, so we can test modification.

  $ summer -db=db.sqlite3 generate .
  0s: 0 matched, 0 modified, 3 new, 0 corrupted

Check handling of new and updated files.

  $ echo trova > nueva
  $ touch empty
  $ summer -n -db=db.sqlite3 verify .
  0s: 1 matched, 2 modified, 1 new, 0 corrupted
  $ summer -n -db=db.sqlite3 update .
  0s: 1 matched, 2 modified, 1 new, 0 corrupted
  $ summer -n -db=db.sqlite3 verify .
  0s: 1 matched, 2 modified, 1 new, 0 corrupted
