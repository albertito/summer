
Tests for how to handle file access issues.
Note we put our test paths in "root/" so the database doesn't
interfere.

  $ alias summer="$TESTDIR/../summer"
  $ mkdir root
  $ touch root/empty
  $ echo marola > root/hola

  $ summer generate root/
  0s: 0 matched, 0 modified, 2 new, 0 corrupted

  $ summer verify root/
  0s: 2 matched, 0 modified, 0 new, 0 corrupted
  $ chmod 0000 root/empty
  $ summer verify root/
  0s: 0 matched, 0 modified, 0 new, 0 corrupted
  open root/empty: permission denied
  [1]

Test behaviour when the root does not exist. This exercises some different
code paths, because the root is special.

  $ summer verify doesnotexist
  0s: 0 matched, 0 modified, 0 new, 0 corrupted
  lstat doesnotexist: no such file or directory
  [1]
